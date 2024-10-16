# Seek: AI-Powered Cover Letter Generator

## Build

`just gotenberg` - starts a docker container required for pdf generation

`just local` - run the program

`just deploy` - deploys to cloud run service by building and pushing latest docker image


## Overview

Seek is a web-based application designed to automate the process of cover letter generation using advanced artificial intelligence technology. It leverages Google's Gemini AI model to produce customized cover letters for job applications.


## Cloud Infrastructure

- Platform: Google Cloud Run
- Benefits:
  - Serverless execution
  - Automatic scaling based on demand
  - Pay-per-use pricing model



## Dependencies 
- [golang 1.23]
- [docker]
- [just]
- [reflex]
- [gcloud cli]


## Resources
- [https://github.com/GoogleContainerTools/distroless](https://github.com/GoogleContainerTools/distroless)
- [https://en.wikipedia.org/wiki/Filesystem_Hierarchy_Standard](https://en.wikipedia.org/wiki/Filesystem_Hierarchy_Standard)
