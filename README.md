# Go Repos

<p align="center">
    <a href="https://gorepos.glup3.dev" target="_blank" rel="noopener">
        <img src="./screenshot.png" alt="Go Repos Screenshot" />
    </a>
</p>

[Go Repos](https://gorepos.glup3.dev), a simple web app that showcases random GitHub repositories written in Go.

## Motivation

I wanted to explore Go projects and improve my Go coding skills.
Using GitHub Search felt slow and clunky, so I built my own UI.
The data is aggregated via the GitHub API and displayed using `net/http` and `html/template`.

> [!WARNING]
> This project is still under heavy development,
> and the codebase has not been cleaned up yet.

## Development

Install Go Dependencies

```sh
go mod tidy
```

Install tailwindcss/cli

```sh
npm install
```

Run main and css watcher

```sh
make api

# in a different terminal
make css-watch
```
