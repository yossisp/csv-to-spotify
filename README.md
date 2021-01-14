![Gitlab pipeline status](https://img.shields.io/gitlab/pipeline/easpex/csv-to-spotify/master)

# README

### About

The project allows you to create playlists in Spotify from a CSV file. It can be used with this frontend [app](https://github.com/yossisp/csv-to-spotify-ui).

### Installation

Execute `make docker-build` in order to build a Docker image with the binary. Running `make docker-run` will create a Docker container and the application can be reached at port 8000 by default. If the client application is a website which will run on a different port make sure that `ALLOWED_ORIGINS` environment variable is updated with the website origin (for [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) reasons).

### Settings

- The complete list of possible environment variables can be found in `.env.example` file. The file should be renamed to `.env` so that all the variables are automatically loaded in the application environment.

- You should set `MARKET` environment variable to the [country](https://developer.spotify.com/documentation/web-api/reference-beta/#category-search) tied to your Spotify account.
- `TEST_REFRESH_TOKEN` is only required for running tests.
- `SPOTIFY_CLIENT_ID_SECRET_BASE64` is of form `<base64 encoded client_id:client_secret>`. You can read more about Spotify authorization [here](https://developer.spotify.com/documentation/general/guides/authorization-guide/#authorization-code-flow).

### Project Overview

The original inspiration for the project was to copy Apple Music playlists to Spotify. This can be done by exporting Itunes playlist as a CSV file (Itunes 12.8) as follows: Itunes -> File -> Library -> Export Playlist. As a result any CSV file formatted to the project-specific format can be uploaded to become a Spotify playlist. I'm aware that there're some solutions on the Internet which can perform the task but most of them are paid and I wasn't satisfied with the one I tried.

The CSV format is:

- The header row with the first column containing song name, the second column containing artist name.
- The file must comma-separated.

The application requires some Spotify user data, most importantly refresh token in order to perform track lookups and add them to user playlist. The application **doesn't collect user email**.

Due to Spotify API rate limiting all playlist tracks can't be looked up at once, instead they're looked up in batches and there's `TRACK_LOOKUP_INTERVAL` seconds in between each batch lookup (`TRACK_LOOKUP_INTERVAL` is an environment variable, by default it's 5 seconds).

The application also has a websocket server which updates client websockets with lookup progress: how many tracks have been found/not found.

The application uses Kafka for messaging between application modules. Inside kafka folder there's a `docker-compose.yml` in order to start a kafka service locally. Personally, I was using [cloudkarafka](https://www.cloudkarafka.com) managed Kafka service which has a free tier.

### Roadmap

- Add more tests.
