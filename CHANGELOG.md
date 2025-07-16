# Changelog
## 2.1.1 (2025-07-16)
Add optional cookie based authentication to the plugin server

Add include cookie credentials for the js plugin

Add a build script to build the plugin for Helm

Add a .gitignore file to ignore the build script and the build output
 
Move docker build to dockerfile @mjibril
  
## 2.1.0 (2021-08-09)

- Adds support for relative URLs in endpoint settings (#23). Thanks to: @mig4

## 2.0.1 (2021-07-26)

- Use Grafana's eslint config (#20).
- Updates plugin dependencies.
- Switches to Grafana plugin workflows for CI/Release.

## 2.0.0 (2021-04-21)

- Panel is now stateless, sessions are matched using a UUID, not a database ID.
- Panel now optionally forwards heartbeats, which can be used to accurately report session duration.
- New payload schema, now including variables & focused state.
- Updates and cleans up plugin dependencies.
- Adds a verified Grafana signature.
- Adds an included backend server (see readme for details).
- Adds better examples, docs, descriptions, etc.

## 1.1.1 (2020-08-06)

- Fixes an issue with Grafana 7.1 compatibility (#5).

## 1.1.0 (2020-08-06)

- Rethrows exceptions on post issues (to make finding non-working panels easier).
- Adds an optional config file that can be used to specify default settings.
- Adds support for template variables (value to change in #6).

## 1.0.0 (2020-06-15)

- Removes references to Grafana Angular components.
- Replaces JSON component with Grafana React JSON component.
- Adds an option to toggle between normal and flattened data.
- Displays errors on the panel when a fetch cannot be completed.

## 0.0.3 (2020-06-09)

- Fixes an issue where references were incorrect in Grafana 7.
- Flattens output JSON to support Telegraf's HTTP listener.
- Adds switchable CORS modes to support some Telegraf environments.

## 0.0.2 (2020-06-09)

- Grafana 7 hotfix.

## 0.0.1 (2019-11-30)

- Initial release.
