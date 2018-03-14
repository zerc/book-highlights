```
DISCLAIMER: the hidden purpose of the project is to try different technologies and at the same time solve the problem of syncrhonisation book highlights between different services (which may not provide API).

With that being said, be aware of the poor implementation ahead. Feel free to contribute though if you like.
```

# Book Highlights

A set of tools to export your book hightlights from different services (Kindle, Google Books etc) to your storage where you can work with them in any way you like e.g. provide API to query them.


**Services supported**

* Amazon Kindle


## Amazon Kindle

Amazon doesn't provide any API to interact with highlights, instead it does provide a website https://read.amazon.co.uk/notebook where you can see them. If you google around you may find a few solutions to parse that page and export the data in different formats.

They didn't work for me because wasn't able to log into my account in order parse the page. I suppose since those packages were developed, Amazon improved its security a lot and nowdays it takes more efforts to pass it through. All solutions I've found are built upon a "mechanize" package (Ruby version) i.e. not real browser. More confusion adds the error message, it says only:

```
Enter a valid email or mobile number
```

Of course I'm 100% that the credentials are correct.

After that I've tried the [headless Chrome](https://hub.docker.com/r/justinribeiro/chrome-headless/) and even if I emulate JS events like `keyup` or `mouse.click`, Amazon somehow detects that something is wrong and asks me for captcha to complete. I don't have much patience to explore the way to fool that protection so I did choose the "semi-automatical" approach:

* I do log in manually, solve the captcha if required and leave the session open
* After that it is possible to run a script to use that session and part the data
