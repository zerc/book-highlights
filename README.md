**DISCLAIMER:** the hidden purpose of the project is to try different technologies and at the same time solve the problem of syncrhonisation book highlights between different services (which may not provide API).

With that being said, be aware of the poor implementation ahead. Feel free to contribute though.

---

# Book Highlights

[![Build Status](https://travis-ci.org/zerc/book-highlights.svg?branch=master)](https://travis-ci.org/zerc/book-highlights) [![Go Report Card](https://goreportcard.com/badge/github.com/zerc/book-highlights)](https://goreportcard.com/report/github.com/zerc/book-highlights) 
[![GitHub license](https://img.shields.io/github/license/zerc/book-highlights.svg)](https://github.com/zerc/book-highlights/blob/master/LICENSE)

This project is a set of tools to export your book hightlights from different services (Kindle, Google Books etc) to your storage where you can work with them in any way you like plus, provide API to query them.


**Services supported**

* [Amazon Kindle](src/kindle/README.md)

### TODO:

* Add a builtin service to store highlights.
* Make the Kindle parser a bit more flexable (debug levels, environment variables).
* Add an example of Ansible playbook to set up everything on the server.
