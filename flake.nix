{
  description = "seshy: multi-repository worktree session manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    # The canonical provider/v1 contract. schema/narrow.cue adds seshy's rules
    # on top of it for the manifest in share/seshy/providers.
    provider-spec = {
      url = "github:roshbhatia/provider-spec/v1.0.0";
      flake = false;
    };
    # The roster release whose catalog schema schema/roster.catalog.v1.schema.json
    # copies. The copy is what the Go tests validate "sy source list" against.
    roster = {
      url = "github:roshbhatia/roster/v0.1.0";
      flake = false;
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      provider-spec,
      roster,
      ...
    }:
    let
      supportedSystems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];
      eachSystem = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      formatter = eachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        pkgs.writeShellApplication {
          name = "seshy-format";
          runtimeInputs = [
            pkgs.fd
            pkgs.nixfmt
          ];
          text = ''exec fd --extension nix --type file --exec-batch nixfmt "$@"'';
        }
      );

      packages = eachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          version =
            let
              root = builtins.readFile ./cmd/root.go;
              match = builtins.match ''.*const version = "([0-9]+\.[0-9]+\.[0-9]+)".*'' root;
            in
            if match == null then
              throw "flake.nix could not read the version from cmd/root.go"
            else
              builtins.head match;
          seshy = pkgs.buildGoModule {
            pname = "seshy";
            inherit version;
            src = ./.;
            vendorHash = "sha256-6ABqBYoJY6AnaX4ThRhrqDBv/ky/h2nAW4R9OYThXFU=";
            subPackages = [ "./cmd/sy" ];
            nativeBuildInputs = [ pkgs.installShellFiles ];
            nativeCheckInputs = [
              pkgs.bash
              pkgs.fish
              pkgs.git
              pkgs.nushell
              pkgs.zsh
            ];
            doCheck = true;
            checkPhase = ''
              runHook preCheck
              go test -race ./...
              bash test/integration/run_tests.sh
              bash test/e2e/run_e2e.sh
              go run ./cmd/sy generate --check
              bash -n completions/sy.bash
              zsh -n completions/sy.zsh
              fish --no-config -n completions/sy.fish
              nu --no-config-file --no-std-lib --commands 'source completions/sy.nu'
              go run ./cmd/sy init bash > "$TMPDIR/sy-init.bash"
              go run ./cmd/sy init zsh > "$TMPDIR/sy-init.zsh"
              go run ./cmd/sy init fish > "$TMPDIR/sy-init.fish"
              go run ./cmd/sy init nu > "$TMPDIR/sy-init.nu"
              bash -n "$TMPDIR/sy-init.bash"
              zsh -n "$TMPDIR/sy-init.zsh"
              fish --no-config -n "$TMPDIR/sy-init.fish"
              nu --no-config-file --no-std-lib --commands "source '$TMPDIR/sy-init.nu'"
              runHook postCheck
            '';
            postInstall = ''
              installShellCompletion \
                --cmd sy \
                --bash <("$out/bin/sy" completion bash) \
                --fish <("$out/bin/sy" completion fish) \
                --zsh <("$out/bin/sy" completion zsh)
              mkdir -p "$out/share/nushell/vendor/autoload"
              "$out/bin/sy" completion nu > "$out/share/nushell/vendor/autoload/sy.nu"
              install -Dm644 share/seshy/providers/seshy.yaml "$out/share/seshy/providers/seshy.yaml"
              ln -s "$out/bin/sy" "$out/bin/seshy"
            '';
            meta = {
              description = "Multi-repository worktree session manager";
              homepage = "https://github.com/roshbhatia/seshy";
              license = pkgs.lib.licenses.mit;
              mainProgram = "sy";
              platforms = pkgs.lib.platforms.unix;
            };
          };
        in
        {
          inherit seshy;
          default = seshy;
          provider-wezterm = import ./extras/wezterm {
            inherit pkgs;
            core = seshy;
          };
          extras = self.packages.${system}.provider-wezterm;
          full = pkgs.symlinkJoin {
            name = "seshy-full";
            paths = [
              seshy
              self.packages.${system}.extras
            ];
            meta = seshy.meta;
          };
        }
      );

      apps = eachSystem (system: {
        default = {
          type = "app";
          program = "${nixpkgs.lib.getExe self.packages.${system}.default}";
        };
      });

      checks = eachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          package = self.packages.${system}.seshy;
        in
        {
          inherit package;
          formatting = pkgs.runCommand "seshy-formatting" { nativeBuildInputs = [ pkgs.nixfmt ]; } ''
            nixfmt --check ${./flake.nix}
            touch "$out"
          '';
          installed-surfaces =
            pkgs.runCommand "seshy-installed-surfaces" { nativeBuildInputs = [ pkgs.nushell ]; }
              ''
                test -x ${package}/bin/sy
                test -L ${package}/bin/seshy
                test -s ${package}/share/bash-completion/completions/sy.bash
                test -s ${package}/share/zsh/site-functions/_sy
                test -s ${package}/share/fish/vendor_completions.d/sy.fish
                test -s ${package}/share/nushell/vendor/autoload/sy.nu
                test -s ${package}/share/seshy/providers/seshy.yaml
                ${package}/bin/sy --version | grep -F 'sy version '
                ${package}/bin/sy init nu > "$TMPDIR/sy-init.nu"
                cmp "$TMPDIR/sy-init.nu" ${package}/share/nushell/vendor/autoload/sy.nu
                test "$(grep -Ec '^export (def --env sy|extern "sy") \[$' ${package}/share/nushell/vendor/autoload/sy.nu)" -eq 1
                export PATH=${package}/bin:$PATH
                export XDG_STATE_HOME="$TMPDIR/state"
                export SESHY_TEST_WORK_PATH="$XDG_STATE_HOME/seshy/sessions/work"
                export SESHY_TEST_LIST_PATH="$XDG_STATE_HOME/seshy/sessions/list"
                mkdir -p "$SESHY_TEST_WORK_PATH" "$SESHY_TEST_LIST_PATH"
                cat > "$TMPDIR/installed.nu" <<'NU'
                source ${package}/share/nushell/vendor/autoload/sy.nu
                let roots = (scope commands | where name == "sy")
                if (($roots | length) != 1) {
                  error make { msg: "installed integration defines sy more than once" }
                }
                let listing = (sy list | complete)
                if ($listing.exit_code != 0) or (not ($listing.stdout | str contains "list")) {
                  error make { msg: "reserved list command did not reach the CLI" }
                }
                sy work
                if $env.PWD != $env.SESHY_TEST_WORK_PATH {
                  error make { msg: "one-word session navigation did not change directory" }
                }
                let resolved = (sy --greedy list | str trim)
                if $resolved != $env.SESHY_TEST_LIST_PATH {
                  error make { msg: "explicit --greedy did not resolve a conflicting session" }
                }
                NU
                nu --no-config-file --no-std-lib "$TMPDIR/installed.nu"
                touch "$out"
              '';
          # The manifest satisfies the spec plus schema/narrow.cue, every negative
          # fixture is rejected, and the committed catalog schema is the roster
          # release the flake pins.
          provider-spec-contract =
            pkgs.runCommand "seshy-provider-spec-contract"
              {
                nativeBuildInputs = [
                  pkgs.cue
                  pkgs.diffutils
                ];
              }
              ''
                cd ${./.}
                export HOME="$TMPDIR"
                cue vet -d '#Manifest' ${provider-spec}/provider.cue schema/narrow.cue share/seshy/providers/seshy.yaml
                for fixture in schema/fixtures/*.yaml; do
                  if cue vet -d '#Manifest' ${provider-spec}/provider.cue schema/narrow.cue "$fixture" 2>/dev/null; then
                    echo "reject expected: $fixture" >&2
                    exit 1
                  fi
                done
                diff -u ${roster}/schema/roster.catalog.v1.schema.json schema/roster.catalog.v1.schema.json
                touch "$out"
              '';
          media-freshness =
            pkgs.runCommand "seshy-media-freshness"
              {
                nativeBuildInputs = [
                  pkgs.imagemagick
                  pkgs.ripgrep
                ];
              }
              ''
                cd ${./.}
                bash ./hack/screenshots.sh --check
                touch "$out"
              '';
        }
      );

      devShells = eachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShellNoCC {
            packages = [
              pkgs.uv
              pkgs.python3
              pkgs.ffmpeg
              pkgs.actionlint
              pkgs.bash
              pkgs.fish
              pkgs.charm-freeze
              pkgs.cue
              pkgs.git
              pkgs.go
              pkgs.go-task
              pkgs.go-tools
              pkgs.golangci-lint
              pkgs.goreleaser
              pkgs.gopls
              pkgs.gotools
              pkgs.imagemagick
              pkgs.nixfmt
              pkgs.nushell
              pkgs.ripgrep
              pkgs.shellcheck
              pkgs.shfmt
              pkgs.vhs
              pkgs.zsh
            ];
            shellHook = ''
              export GOTOOLCHAIN=local
            '';
          };
        }
      );
    };
}
