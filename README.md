# gore-bot

`gore-bot` (tentative name) is a bot designed to listen to a filtered Twitter stream for a specified Twitter handle to watch for game captures posted from PS4 or Switch so it can push the attached captures to a defined Amazon S3 bucket.

## Usage

There are two pieces of setup required:

* Make a copy of the `.env-sample` file as `.env` and update the values with the appropriate credentials for your Twitter app and AWS S3 account info (or populate the values as environment variables)
* Make a copy of the `config-sample.toml` file as `config.toml` and update the various values as desired

Once setup is complete, simply run the application and it will attach to the Twitter stream to react to matching tweets.

## Planned work

* [ ] Add basic validity check for credential struct(s) to bail out ASAP e.g. if any values are empty
* [ ] Finish code to process tweets including parsing out relevant metadata
* [ ] Finish code to take results from processing tweets and get media to S3
* [ ] Improve flexibility of required data and CLI options (e.g. location of config files)
* [ ] Add a license
* [ ] Unit tests! >_>
