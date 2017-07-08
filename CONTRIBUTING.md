# Contributing

We welcome pull requests from everyone. By participating in this project, you agree to abide by the
[code of conduct](CODE_OF_CONDUCT.md).

## Forking the Code

Here's a [blog post](http://blog.campoy.cat/2014/03/github-and-go-forking-pull-requests-and.html) on how to fork a go
project.

Building tstask in go (assuming you're up and running with a go already) is as simple as:

```
go get github.com/trafero/tstack/cmd/...
```

## What to Do Before Making a Pull Request

* Please use "go fmt" to format your code
* Please check that the [Paho functional tests](https://github.com/trafero/tstack/blob/master/docs/performance.md#functional-testing)
pass before submitting a pull request

## What Happens After You Create a Pull Request?

At this point you're waiting on us. We like to at least comment on pull requests within three business days. 
We may suggest some changes or improvements or alternatives. One thing is for sure - we'll be delighted that you have
taken the time to support the project.

Some things that will increase the chance that your pull request is accepted:

* Write tests
* Comment your code
* Write a good commit message
