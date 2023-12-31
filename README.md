# Description
json2steamshortcut is used to generate a shortcut.vdf (the file that contains all the "non-steam-games") from a json.

# Intention
json2steamshortcut is intended to be used with Nix/NixOS to enable declaratively definining the links in steam. 
I personally use it in combination with [Jovian-NixOS](https://github.com/Jovian-Experiments/Jovian-NixOS) to get declaratively a gameconsole like system without needing to fall back to a WM/DE just for setting up shortcuts in steam.

# Usage Example
```
{
  config,
  lib,
  pkgs,
  ...
}: let
  # Declare our shortcuts natively in our nix configuration
  json = builtins.toJSON [
    {
      "AppName" = "Performous";
      "ExePath" = "${pkgs.performous}/bin/performous";
    }
  ];
  # Create the vdf file at build time
  vdf = pkgs.runCommand "shortcuts.vdf" {
    nativeBuildInputs = [pkgs.json2steamshortcut];
  } "echo '${json}' | json2steamshortcut > $out";
in {

  programs.steam.enable = true;

  # use homemanager to place our shortcuts.vdf at the correct location (this is user and steamaccount specific)
  home-manager.users.chris.home = {
    file.".local/share/Steam/userdata/158842264/config/shortcuts.vdf".source = vdf;
  };
}
```

# Supported Fields
This tool is using the [steamutil](https://github.com/stephen-fox/steamutil) library for vdf generation. We support all the fields this lib supports.
Currently these are this ones:
```
type Shortcut struct {
	Id                 int
	AppName            string
	ExePath            string
	StartDir           string
	IconPath           string
	ShortcutPath       string
	LaunchOptions      string
	IsHidden           bool
	AllowDesktopConfig bool
	AllowOverlay       bool
	IsOpenVr           bool
	LastPlayTimeEpoch  int32
	Tags               []string
}
```

# TODO
- Create a proper home-manager module

# Known Bugs
- Steam will overwrite this files
	- Software added through steam will be lost with the next update.
   	- Home Manager will complain after the second usage, since it will backup the original file and when aready a backup exists refuse to work. I haven't found a way to disable this behaviour.
    
A possible solution would be to add merging support and do this during runtime (and not build-time). So the declarative json will be merged with a pre-existing shortcuts.vdf when home-manager applies it's config. Currently i'm fine with these bugs, so i propably won't fix them. But PRs are welcome if someone wants to do this. 
