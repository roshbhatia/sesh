# wezterm

Select a checkout group for token parser review.

## Install

```sh
brew install roshbhatia/tap/seshy-provider-wezterm
nix profile add 'github:roshbhatia/seshy#provider-wezterm'
```

Install the core utility separately, or select its all-provider bundle. Runtime tools still need their own credentials.

## Demo

![Select a checkout group for token parser review](demo.gif)

[Tape source](demo.tape) · [Task script](demo.sh)

Run `nix develop -c bash extras/wezterm/demo.sh` to run the task without recording.
Run `nix develop -c python3 hack/extra-demos.py wezterm` to record it.
