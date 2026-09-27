---
title: Customize your theme
weight: 111
aliases:
  - /strategies/themes/
  - /themes/
---

Customize the waiting page shown while an instance starts, or the error page for a missing group or theme.

```yaml
services:
  sablier:
    image: sablierapp/sablier:{{< version >}}
    restart: always
    volumes:
      - '/var/run/docker.sock:/var/run/docker.sock'
      - '/path/to/my/themes:/etc/sablier/themes'
```

Sablier comes with a set of default themes that you can use.

You can also extend the themes by providing your own, which will be rendered as Go Templates.

## The embedded themes


|       Name        |                                                  Preview                                                  |
| :---------------: | :-------------------------------------------------------------------------------------------------------: |
|      `ghost`      |           ![ghost](/assets/img/ghost.png)           |
|     `shuffle`     |         ![shuffle](/assets/img/shuffle.png)         |
| `hacker-terminal` | ![hacker-terminal](/assets/img/hacker-terminal.png) |
|     `matrix`      |          ![matrix](/assets/img/matrix.png)          |
|      `retro`      |           ![retro](/assets/img/retro.png)           |



## Custom themes locations

`--strategy.dynamic.custom-themes-path` defaults to `/etc/sablier/themes`, so the volume mount
shown at the top of this page is all you need: no extra argument is required. Set the argument
only when you want Sablier to look somewhere else.

If the folder does not exist, Sablier serves the embedded themes and says so at `debug` level,
so leaving the default in place costs nothing when you have no custom themes.

Sablier will recursively search for themes with the `.html` extension. Error templates use the same directory and asset bundling rules as loading themes.

- You **cannot** load new themes added to the folder without restarting
- You **can** modify existing theme files without restarting

## Asset bundling

Sablier bundles external assets at startup when loading custom themes. Any **relative** CSS, JavaScript, or image reference inside a custom `.html` file is read from the same directory tree and inlined directly into the template, so the browser receives a fully self-contained page with no additional requests.

| HTML pattern | What Sablier inlines |
|---|---|
| `<link rel="stylesheet" href="css/style.css">` | `<style>/* contents of css/style.css */</style>` |
| `<script src="js/app.js"></script>` | `<script>/* contents of js/app.js */</script>` |
| `<img src="imgs/logo.png">` | `<img src="data:image/png;base64,...">` |

Paths that are **not** inlined and remain unchanged:

- Absolute URLs: `https://cdn.example.com/style.css`
- Protocol-relative URLs: `//cdn.example.com/style.css`
- Root-relative paths: `/static/style.css`
- Data URIs already present in the HTML: `data:image/png;base64,...`

If a referenced file cannot be found on disk the original tag is kept as-is, so a missing asset causes a broken browser request rather than a startup error.

### Directory layout example

```
/path/to/my/strategies/themes/
├── my-theme.html
├── css/
│   └── style.css
└── imgs/
    ├── logo.png
    └── favicon.png
```

```html
<!-- my-theme.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta http-equiv="refresh" content="{{ .RefreshFrequency }}" />
  <link rel="stylesheet" href="css/style.css">
</head>
<body>
  <img src="imgs/logo.png" alt="Logo">
  <p>Starting {{ .DisplayName }}...</p>
  <script src="js/app.js"></script>
</body>
</html>
```

At startup Sablier reads `css/style.css`, `imgs/logo.png`, and `js/app.js` relative to `my-theme.html` and produces a single self-contained template. External CDN links (e.g. Google Fonts) are left untouched.

## Create a custom theme

Themes are served using [Go Templates](https://pkg.go.dev/text/template).

### Available Go template values

| Template Key                                  | Template Value                                                                                                      | Go Template Usage                                                                        |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `.DisplayName`                                | The display name configured for the session                                                                         | `{{ .DisplayName }}`                                                                     |
| `.InstanceStates`                             | An array of `RenderOptionsInstanceState` that represents the state of each required instances                       | `{{- range $i, $instance := .InstanceStates }}{{ end -}}`                                |
| `.SessionDuration`                            | The humanized session duration from a [time.Duration](https://pkg.go.dev/time#Duration)                             | `{{ .SessionDuration }}`                                                                 |
| `.RefreshFrequency`                           | The refresh frequency for the page. See [The `<meta http-equiv="refresh" />` tag](#the-meta-http-equivrefresh--tag) | `<meta http-equiv="refresh" content="{{ .RefreshFrequency }}" />`                        |
| `.Version`                                    | Sablier version as a string                                                                                         | `{{ .Version }}`                                                                         |
| `$RenderOptionsInstanceState.Name`            | The name of the instance loading                                                                                    | `{{- range $i, $instance := .InstanceStates }}{{ $instance.Name }}{{ end -}}`            |
| `$RenderOptionsInstanceState.CurrentReplicas` | The number of current replicas of the instance loading                                                              | `{{- range $i, $instance := .InstanceStates }}{{ $instance.CurrentReplicas }}{{ end -}}` |
| `$RenderOptionsInstanceState.DesiredReplicas` | The number of desired replicas of the instance loading                                                              | `{{- range $i, $instance := .InstanceStates }}{{ $instance.DesiredReplicas }}{{ end -}}` |
| `$RenderOptionsInstanceState.Status`          | The status of the instance loading, `ready` or `not-ready`                                                          | `{{- range $i, $instance := .InstanceStates }}{{ $instance.Status }}{{ end -}}`          |
| `$RenderOptionsInstanceState.Error`           | The error trigger by this instance which won't be able to load                                                      | `{{- range $i, $instance := .InstanceStates }}{{ $instance.Error }}{{ end -}}`           |

### The `<meta http-equiv="refresh" />` tag

The auto-refresh is accomplished using the [HTML <meta> http-equiv Attribute](https://www.w3schools.com/tags/att_meta_http_equiv.asp).

> Defines a time interval for the document to refresh itself.

The first step to creating your own theme is to include the `HTML <meta> http-equiv Attribute` as follows:

```html
<head>
  ...
  <meta http-equiv="refresh" content="{{ .RefreshFrequency }}" />
  ...
</head>
```

## Customize dynamic error pages

The dynamic strategy can show an HTML page when a group or theme is missing. It uses the selected theme's error template: `ghost`, for example, uses `ghost.error.html`. To customize yours, add a matching file in Sablier's custom themes directory:

```text
/path/to/my/themes/
├── my-theme.html         # waiting page: theme=my-theme
└── my-theme.error.html   # error page for theme=my-theme
```

This also works in subdirectories. If the theme or its error template is missing, Sablier uses `error.html`. Add your own `error.html` to replace that fallback. Error templates don't appear in `/api/themes` and can't be selected as waiting pages.

For example, `my-theme.error.html` can contain:

```html
<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>{{ .Title }}</title></head>
<body>
  <h1>{{ .StatusCode }} {{ .StatusText }}</h1>
  <p>{{ .Detail }}</p>
  {{- if .RequestedGroup }}<p>Requested group: {{ .RequestedGroup }}</p>{{ end }}
  {{- if .AvailableGroups }}
  <p>Available groups:</p>
  <ul>{{ range .AvailableGroups }}<li>{{ . }}</li>{{ end }}</ul>
  {{- end }}
</body>
</html>
```

Templates use [Go `html/template`](https://pkg.go.dev/html/template), which escapes values automatically. Available fields are:

- `.StatusCode`, `.StatusText`, `.Title`, `.Detail`, `.Version`
- `.RequestedGroup`, `.AvailableGroups` for a missing group
- `.RequestedTheme`, `.AvailableThemes` for a missing theme

Waiting-page fields such as `.DisplayName` and `.RefreshFrequency` aren't available. Leave out the auto-refresh tag: retrying won't fix a missing group or theme.

Sablier returns HTML when `Accept` prefers `text/html` to JSON. Otherwise it keeps the JSON response, including for missing headers, `*/*`, ties, or XHTML-only requests. Validation and internal errors still return JSON.

Try both formats with a nonexistent group:

```bash
curl -i -H 'Accept: text/html' 'http://localhost:10000/api/strategies/dynamic?group=does-not-exist&theme=ghost'
curl -i -H 'Accept: application/problem+json' 'http://localhost:10000/api/strategies/dynamic?group=does-not-exist&theme=ghost'
```

Both return 404 with `Vary: Accept`: the first as HTML, the second as `application/problem+json`.

### Using a reverse proxy

The plugin must forward the browser's `Accept` header to Sablier and relay the error status, content type, body, and cache headers back to the browser. Selecting a theme alone isn't enough.

Caddy, Traefik, and Proxy-Wasm need companion plugin changes for this feature; updating Sablier alone won't enable it. Check your plugin's support, then repeat the requests above against your application URL. You should get the same 404 and content type, without the request reaching the backend.

## The `showDetails` option

If `showDetails` is set to `false`, the `.InstanceStates` will be an empty array.

## How to load your custom theme

You can load themes by specifying their name and relative path from the `--strategy.dynamic.custom-themes-path` value.

```bash
/my/custom/strategies/themes/
├── custom1.html      # custom1
├── custom2.html      # custom2
└── special
    └── secret.html   # special/secret
```

For example:

```bash
curl 'http://localhost:10000/api/strategies/dynamic?session_duration=1m&names=nginx&theme=custom1'
```

## See the available themes from the API

The themes endpoint is served at `/api/themes` (the legacy path `/api/dynamic/themes` still
works). It returns every loaded theme, embedded and custom alike, in a single sorted list.

```bash
curl 'http://localhost:10000/api/themes'
```

With the layout shown in [How to load your custom theme](#how-to-load-your-custom-theme):

```json
{
  "themes": [
    "custom1",
    "custom2",
    "ghost",
    "hacker-terminal",
    "matrix",
    "retro",
    "shuffle",
    "special/secret"
  ]
}
```

See the [API reference](/reference/api/) for the full `themes` endpoint documentation.

## Design your theme in the browser

To build and preview a theme visually, use the **[Sablier theme editor](https://editor.sablierapp.dev)**, then export the HTML into your themes folder.