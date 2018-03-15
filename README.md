**DISCLAIMER:** the hidden purpose of the project is to try different technologies and at the same time solve the problem of syncrhonisation book highlights between different services (which may not provide API).

With that being said, be aware of the poor implementation ahead. Feel free to contribute though.

---

# Book Highlights

[![Build Status](https://travis-ci.org/zerc/book-highlights.svg?branch=master)](https://travis-ci.org/zerc/book-highlights) [![Go Report Card](https://goreportcard.com/badge/github.com/zerc/book-highlights)](https://goreportcard.com/report/github.com/zerc/book-highlights) 
[![GitHub license](https://img.shields.io/github/license/zerc/book-highlights.svg)](https://github.com/zerc/book-highlights/blob/master/LICENSE)

This project is a set of tools to export your book hightlights from different services (Kindle, Google Books etc) to your storage where you can work with them in any way you like plus, provide API to query them.


**Services supported**

* [Amazon Kindle](src/kindle/README.md)

# Quick start

You can either build everything by yourself (see README files for each service for details) or use images from DockerHub.

Here is an example of the compose file to use pre built images:

```yml
version: '3'
services:
  kindle:
    image: "zerc/book-highlights-kindle"
    depends_on:
     - chrome
    environment:
     - API_ENTRYPOINT=API_URL
     - CHROME_DEBUG=0
  chrome:
    image: "justinribeiro/chrome-headless"
    ports:
      - "9222:9222"
    cap_add:
      - SYS_ADMIN
```

Where `API_URL` needs to be replaced to your API endpoint to accept the payload with highlights parsed.


### TODO:

* Add a builtin service to store highlights.
* Make the Kindle parser a bit more flexable (debug levels, environment variables).
* Add an example of Ansible playbook to set up everything on the server.
