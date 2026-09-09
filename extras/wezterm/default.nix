{ pkgs, core }:
let
  executable = core.overrideAttrs (old: {
    pname = "seshy-picker";
    subPackages = [ "extras/wezterm" ];
    nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ pkgs.makeWrapper ];
    checkPhase = ''
      runHook preCheck
      go test -race ./extras/wezterm
      runHook postCheck
    '';
    postInstall = ''
      mv "$out/bin/wezterm" "$out/bin/seshy-picker"
      wrapProgram "$out/bin/seshy-picker" --add-flags "${pkgs.lib.getExe core}"
    '';
    meta = old.meta // {
      mainProgram = "seshy-picker";
    };
  });
  manifest = pkgs.writeText "seshy.json" (
    builtins.toJSON (
      (builtins.fromJSON (builtins.readFile ./provider.json))
      // {
        command = [ (pkgs.lib.getExe executable) ];
      }
    )
  );
in
pkgs.symlinkJoin {
  name = "seshy-picker";
  paths = [ executable ];
  postBuild = "mkdir -p $out/share/wezterm/providers; cp ${manifest} $out/share/wezterm/providers/seshy.json";
  meta.mainProgram = "seshy-picker";
}
