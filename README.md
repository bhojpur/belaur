# Bhojpur Belaur - Pipeline Management

The `Bhojpur Belaur` is a *high performance* __DevOps__ *pipeline management* engine
applied within the [Bhojpur.NET Platform](https://github.com/bhojpur/platform)
ecosystem for delivery of distributed `applications` or `services`. Initially, it
was designed for software builds, but it could be customised for `AI/ML` models or
to process any kind of data (e.g., image, audio/video, simulation models, CRM/ERP).

## Getting Started

Firstly, the `Bhojpur Belaur` leverages a standard `Kubernetes` instance to manage
multiple build pipelines for your software applications.

### Build Source Code

Firstly, clone this `Git` reporitory in a folder. You would require `Go` programming
language runtime to build the `Belaur CLI`. You need `Node.js` also for building the
`Belaur Dashboard`. Now, issue the following commands in a terminal window.

```bash
go mod tidy
make compile_frontend
make static_assets
make compile_backend
```

if successful, then you could run the following command in another terminal window.

```bash
make dev
```

or

```bash
belaur -home-path=${PWD}/tmp -dev=true
```

in a development mode. Now, open the `http://localhost:8080` URL in a web broswer.

### Installation

The installation of `Bhojpur Belaur` is simple and often takes a few minutes only.

#### Using Docker

The following command starts `Bhojpur Belaur` as a daemon process and mounts all data
to the current folder. Afterwards, `Bhojpur Belaur` will be available on the host system
on port 8080. Use the standard user *admin* and password *admin* as initial login. It is
recommended to change the password afterwards.

```bash
docker run -d -p 8080:8080 -v $PWD:/data bhojpur/belaur:latest
```

It uses the `Docker` image with the *latest* tag, which includes all required libraries
and compilers for all supported languages. If you prefer a smaller image suited for a
preferred language, then you have a look at the `available docker image tags`_.

#### Manually

It is possible to install `Bhojpur Belaur` directly on the host system. This can be
achieved by downloading the binary from the `releases page`_.

The `Bhojpur Belaur` will automatically detect the folder of the binary and will
place all data next to it. You can change the data directory with the startup
parameter *-home-path*, if you want.

#### Using Helm

If you haven't got an `ingress controller` pod yet, make sure that you have
`kube-dns` or `coredns` enabled, run this command to set it up.

```bash
make kube-ingress
```

To init helm:

```bash
helm init
```

To deploy `Bhojpur Belaur`:

```bash
make deploy-kube
```

### Example Pipelines

#### Go

```golang
    package main

    import (
        "log"

        sdk "github.com/bhojpur/gosdk"
    )

    // This is one job. Add more if you want.
    func DoSomethingAwesome(args sdk.Arguments) error {
        log.Println("This output will be streamed back to Bhojpur Belaur and will be displayed in the pipeline logs.")

        // An error occurred? Return it back so Bhojpur Belaur knows that this job failed.
        return nil
    }

    func main() {
        jobs := sdk.Jobs{
            sdk.Job{
                Handler:     DoSomethingAwesome,
	            Title:       "DoSomethingAwesome",
		        Description: "This job does something awesome.",
	        },
	    }

	    // Serve
	    if err := sdk.Serve(jobs); err != nil {
	        panic(err)
	    }
    }
```

#### Python

```python
    from belaursdk import sdk
    import logging

    def MyAwesomeJob(args):
        logging.info("This output will be streamed back to Bhojpur Belaur and will be displayed in the pipeline logs.")
        # Just raise an exception to tell Bhojpur Belaur if a job failed.
        # raise Exception("Oh no, this job failed!")

    def main():
        logging.basicConfig(level=logging.INFO)
        myjob = sdk.Job("MyAwesomeJob", "Do something awesome", MyAwesomeJob)
        sdk.serve([myjob])
```

#### Java

```java
    package net.bhojpur.belaur;

    import net.bhojpur.belaur.sdk.*;
    import java.util.ArrayList;
    import java.util.Arrays;
    import java.util.logging.Logger;

    public class Pipeline
    {
        private static final Logger LOGGER = Logger.getLogger(Pipeline.class.getName());

        private static Handler MyAwesomeJob = (belaurArgs) -> {
            LOGGER.info("This output will be streamed back to Bhojpur Belaur and will be displayed in the pipeline logs.");
	    // Just raise an exception to tell Bhojpur Belaur if a job failed.
            // throw new IllegalArgumentException("Oh no, this job failed!");
        };

        public static void main( String[] args )
        {
            PipelineJob myjob = new PipelineJob();
            myjob.setTitle("MyAwesomeJob");
            myjob.setDescription("Do something awesome.");
            myjob.setHandler(MyAwesomeJob);

            Javasdk sdk = new Javasdk();
            try {
                sdk.Serve(new ArrayList<>(Arrays.asList(myjob)));
            } catch (Exception ex) {
                ex.printStackTrace();
            }
        }
    }
```

#### C++

```cpp
   #include "cppsdk/sdk.h"
   #include <list>
   #include <iostream>

   void DoSomethingAwesome(std::list<belaur::argument> args) throw(std::string) {
      std::cerr << "This output will be streamed back to Bhojpur Belaur and will be displayed in the pipeline logs." << std::endl;

      // An error occurred? Return it back so Bhojpur Belaur knows that this job failed.
      // throw "Uhh something badly happened!"
   }

   int main() {
      std::list<belaur::job> jobs;
      belaur::job awesomejob;
      awesomejob.handler = &DoSomethingAwesome;
      awesomejob.title = "DoSomethingAwesome";
      awesomejob.description = "This job does something awesome.";
      jobs.push_back(awesomejob);

      try {
         belaur::Serve(jobs);
      } catch (string e) {
         std::cerr << "Error: " << e << std::endl;
      }
   }
```

#### Ruby

```ruby
   require 'rubysdk'

   class Main
       AwesomeJob = lambda do |args|
           STDERR.puts "This output will be streamed back to Bhojpur Belaur and will be displayed in the pipeline logs."

           # An error occurred? Raise an exception and Bhojpur Belaur will fail the pipeline.
           # raise "Oh gosh! Something went wrong!"
       end

       def self.main
           awesomejob = Interface::Job.new(title: "Awesome Job",
                                           handler: AwesomeJob,
                                           desc: "This job does something awesome.")

           begin
               RubySDK.Serve([awesomejob])
           rescue => e
               puts "Error occured: #{e}"
               exit(false)
           end
       end
   end
```

#### Node.js

```javascript
   const nodesdk = require('@bhojpur/belaursdk');

   function DoSomethingAwesome(args) {
       console.error('This output will be streamed back to Bhojpur Belaur and will be displayed in the pipeline logs.');

       // An error occurred? Throw it back so Bhojpur Belaur knows that this job failed.
       // throw new Error('My error message');
   }

   // Serve
   try {
       nodesdk.Serve([{
           handler: DoSomethingAwesome,
           title: 'DoSomethingAwesome',
           description: 'This job does something awesome.'
       }]);
   } catch (err) {
       console.error(err);
   }
```

Pipelines are defined by jobs and a function usually represents a job. You can define as
many jobs in your pipeline as you want.

Every function accepts arguments. Those arguments can be requested from the  itself and
the values are passed back in from the UI.

Some pipeline jobs need a specific order of execution. `DependsOn` allows you to declare
dependencies for every job.
