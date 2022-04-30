package pipelines

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/pipelinehelper"
	"github.com/bhojpur/belaur/pkg/services"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"
)

// Hook represent a github based webhook context.
type Hook struct {
	Signature string
	Event     string
	ID        string
	Payload   []byte
}

// Repository contains information about the repository. All we care about
// here are the possible urls for identification.
type Repository struct {
	GitURL  string `json:"git_url"`
	SSHURL  string `json:"ssh_url"`
	HTMLURL string `json:"html_url"`
}

// Payload contains information about the event like, user, commit id and so on.
// All we care about for the sake of identification is the repository.
type Payload struct {
	Repo Repository `json:"repository"`
}

func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	_, _ = computed.Write(body)
	return []byte(computed.Sum(nil))
}

func verifySignature(secret []byte, signature string, body []byte) bool {
	signaturePrefix := "sha1="
	signatureLength := 45

	if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
		return false
	}

	actual := make([]byte, 20)
	_, _ = hex.Decode(actual, []byte(signature[5:]))
	expected := signBody(secret, body)
	return hmac.Equal(expected, actual)
}

func checkHeaders(req *http.Request) (Hook, error) {
	h := Hook{}

	if h.Signature = req.Header.Get("x-hub-signature"); len(h.Signature) == 0 {
		return Hook{}, errors.New("no signature")
	}

	if h.Event = req.Header.Get("x-github-event"); len(h.Event) == 0 {
		return Hook{}, errors.New("no event")
	}

	if h.Event != "push" {
		if h.Event == "ping" {
			return Hook{Event: "ping"}, nil
		}
		return Hook{}, errors.New("invalid event")
	}

	if h.ID = req.Header.Get("x-github-delivery"); len(h.ID) == 0 {
		return Hook{}, errors.New("no event id")
	}
	return h, nil
}

// GitWebHook handles callbacks from GitHub's webhook system.
// @Summary Handle github webhook callbacks.
// @Description This is the global endpoint which will handle all github webhook callbacks.
// @Tags pipelines
// @Accept json
// @Produce plain
// @Param payload body Payload true "A github webhook payload"
// @Success 200 {string} string "successfully processed event"
// @Failure 400 {string} string "Bind error and schedule errors"
// @Failure 500 {string} string "Various internal errors running and triggering and rebuilding pipelines. Please check the logs for more information."
// @Router /pipeline/githook [post]
func (pp *PipelineProvider) GitWebHook(c echo.Context) error {
	vault, err := services.DefaultVaultService()
	if err != nil {
		return c.String(http.StatusInternalServerError, "unable to initialize vault: "+err.Error())
	}

	err = vault.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, "unable to open vault: "+err.Error())
	}

	req := c.Request()
	req.Header.Set("Content-type", "application/json")
	defer req.Body.Close()

	h, err := checkHeaders(req)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if h.Event == "ping" {
		return c.NoContent(http.StatusOK)
	}

	p := Payload{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	h.Payload = body
	if err := json.Unmarshal(h.Payload, &p); err != nil {
		return c.String(http.StatusBadRequest, "error in unmarshalling json payload")
	}
	var foundPipeline *belaur.Pipeline
	for _, pipe := range pipeline.GlobalActivePipelines.GetAll() {
		if pipe.Repo.URL == p.Repo.GitURL || pipe.Repo.URL == p.Repo.HTMLURL || pipe.Repo.URL == p.Repo.SSHURL {
			foundPipeline = &pipe
			break
		}
	}
	if foundPipeline == nil {
		return c.String(http.StatusInternalServerError, "pipeline not found")
	}
	id := strconv.Itoa(foundPipeline.ID)
	secret, err := vault.Get(belaur.SecretNamePrefix + id)
	migrate := false
	if err != nil {
		// Backwards compatibility, check if there is a secret using the old name.
		secret, err = vault.Get(belaur.LegacySecretName)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		migrate = true
		err = nil
	}
	if !verifySignature(secret, h.Signature, h.Payload) {
		return c.String(http.StatusBadRequest, "invalid signature")
	}

	if migrate {
		// Migrate the secret to a new value using the new format after verification of the signature succeeded.
		// We don't want to migrate an incorrect secret.
		vault.Add(belaur.SecretNamePrefix+id, secret)
		if err := vault.SaveSecrets(); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
	}

	uniqueFolder, err := pipelinehelper.GetLocalDestinationForPipeline(*foundPipeline)
	if err != nil {
		belaur.Cfg.Logger.Error("Pipeline type invalid", "type", foundPipeline.Type)
		return c.String(http.StatusInternalServerError, "pipeline type invalid")
	}
	foundPipeline.Repo.LocalDest = uniqueFolder
	err = pp.deps.PipelineService.UpdateRepository(foundPipeline)
	if err != nil {
		message := fmt.Sprintln("failed to build pipeline: ", err.Error())
		return c.String(http.StatusInternalServerError, message)
	}
	return c.String(http.StatusOK, "successfully processed event")
}
