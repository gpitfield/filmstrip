### filmstrip
filmstrip is a fast, opinionated, minimal, responsive, static photography site generator written in Go. You simply point its config file to a local directory of images, which are converted to a site with the same folder structure as the source directory, with images cut to multiple sizes, and gallery and detail html pages generated from EXIF and related metadata. It includes a driver for upload to S3 (just add your AWS creds to the config file), so you can host your site essentially for free, or pennies.

Pull requests welcome!

#### Setup
First, you'll need to install [Go](https://golang.org/doc/install).
Next, run `go get -u github.com/gpitfield/filmstrip`, which will copy the filmstrip repo to your `GOPATH`.
At the top level of the filmstrip directory you will find the `config.yml` file. Edit it per the instructions below to customize your filmstrip site.
Once that's done, run `go run main.go build` to generate your site, and `go run main.go deploy` to push it to S3.

#### Lightroom EXIF options
Though it's not required, filmstrip is meant to work with Lightroom. If you export a file from Lightroom, you can tell Lightroom to run filmstrip after the image is saved and it will automatically update your site. The best way to do this is to build filmstrip via `go build .` in the `GOPATH` filmstrip directory, and then tell Lightroom to run that binary on export. In addition to the obvious ones to do with camera settings, filmstrip makes use of the "Caption" field in Lightroom to generate image descriptions.

#### filmstrip Directives
 - --force forces filmstrip to rebuild all HTML files, even for images that haven't changed. This can be useful when fiddling with different config options

#### Config Options
 - source-dir: the full path to the local directory of images filmstrip should use to generate the site from.
 - site-title: the name to show on the left side of the home navigation, as well as the page title.
 - copyright: You can specify a default copyright attribution using the `copyright` config value. It will be used for images that do not have EXIF copyright data.
 - cover-columns: the number of image columns to use on the home page, or any other page that is a collection of galleries (e.g. in the case of sub-collections).
 - gallery-columns: the number of image columns to use on a gallery page.
 - about-headline: the headline to show on the about page.
 - about-text: a list of paragraphs to include as the text on the about page.
 - about-image: the full local path to the image to use on the about page.
 - s3-bucket: the name of the s3 bucket to use for the site
 - s3-region: the s3 region to use for the site
 - aws-profile: the aws account profile to use
 - auto-untitle: whether to replace raw camera file names with "Untitled" as their title

#### Images Directory Structure

Image files within the source directory can be organized into directories, and Filmstrip will generate a "collection" for each directory, indexing it on the main navigation. Images can be sorted by using the prefix `_#_`, so for example _2_portrait.jpg will be given index value 2.

The file name, stripped of any sorting prefix and extension, are used as image titles in the generated HTML.

#### Responsiveness
Images are resized in decrements of half from their image size until their next largest dimension would be less than 100px. They are given extensions as _2, _4, _8 which indicates which fraction of the original each image is.

### Test coverage
Derp. There is no test coverage. I built this for my own personal use and while I will probably add some tests, as of right now there aren't any. Boo. I know. Pull request?