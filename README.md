# Gee Bee
An app that alerts you if it detects airplane tail numbers in flight!

### Quick notes
- This app uses quite a bit of code and logic from [Jetspotter](https://github.com/vvanouytsel/jetspotter). Many thanks to [Vincent](https://github.com/vvanouytsel) for his wonderful work there!
- There is an example config file included that you can modify for your own needs. Make a copy of it and name it `config.yml` or `config.yaml`. Sensible defaults are included the app if you'd like to testdrive it first.

# Taskfile tasks
- `build`: build the app as an executable
- `format`: format all of the code to Golang standards
- `run`: run the app

For instance, `task run` will run the app and pull config values from a `config.yml`, `config.yaml`, or environment variables.
