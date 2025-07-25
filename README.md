# json2steamshortcut

## Description

json2steamshortcut is used to generate a shortcut.vdf (the file that contains all the "non-steam-games") from a json.

This tool is using the [steamutil](https://github.com/stephen-fox/steamutil) library for vdf generation.

This repository also provides a NixOS Home-manager module `steam-shortcuts` which provides `services.steam-shortcuts` for creating declarative Steam Shortcuts from your Home-manager configuration.

## Intention

json2steamshortcut is intended to be used with Nix/NixOS to enable declaratively defining the links in steam.
I personally use it in combination with [Jovian-NixOS](https://github.com/Jovian-Experiments/Jovian-NixOS) to get declaratively a gameconsole like system without needing to fall back to a WM/DE just for setting up shortcuts in steam.

## Usage

### CLI

```bash
# Read from stdin
echo '[{"AppName": "MyApp", "Exe": "/path/to/app"}]' | json2steamshortcut > shortcuts.vdf

# Read from file
json2steamshortcut shortcuts.json > shortcuts.vdf

# Or
cat shortcuts.json | json2steamshortcut > shortcuts.vdf
```

### JSON Format

The tool expects valid JSON that contains an array of Shortcut Objects
<details>

<summary>Valid JSON input</summary>

```json
[
  {
    "AppName": "Performous",
    "Exe": "/usr/bin/performous",
    "StartDir": "/home/user",
    "LaunchOptions": "--fullscreen",
    "Icon": "/path/to/icon.png",
    "Tags": ["Music", "Game"]
  }
]
```

</details>

<details>

<summary>Shortcut JSON fields</summary>

- AppName (required): Display name in Steam
- Exe (required): Path to the executable
- StartDir (optional): Working directory
- LaunchOptions (optional): Command line arguments
- Icon (optional): Path to icon file
- Tags (optional): Array of tag strings

</details>

## Home Manager Module

This project provides a home-manager module that makes it easy to declaratively manage Steam shortcuts.

<details>

<summary>Flake inputs</summary>

Add json2steamshortcut to your flake inputs:

```nix
{
  inputs = {
    # ... your other inputs
    json2steamshortcut = {
      url = "github:/ChrisOboe/json2steamshortcut";
      inputs = { 
        nixpkgs.follows = "nixpkgs";
        flake-utils.follows = "flake-utils";
      };
    };
  };
}
```

</details>

<details>

<summary>Home Configuration</summary>

```nix
{
  config,
  lib,
  pkgs,
  inputs,
  ...
}: {
  # Import the home-manager module
  imports = [
    inputs.json2steamshortcut.homeModules.default
  ];

  # Enable Steam
  programs.steam.enable = true;

  # Configure Steam shortcuts using the module
  home-manager.users."user" = {
    services.steam-shortcuts = {
      enable = true;
      overwriteExisting = true; # Recommended for true declarative steam shortcuts - will overwrite any existing shortcuts.vdf
      steamUserId = 158842264; # Replace with your Steam user ID
      shortcuts = [
        {
          AppName = "Performous";
          Exe = "${pkgs.performous}/bin/performous";
        }
        {
          AppName = "Firefox";
          Exe = "${pkgs.firefox}/bin/firefox";
          LaunchOptions = "--new-window";
        }
        {
          AppName = "Lutris";

          # Using Executable name directly is also valid as long as you expect `lutris` to be available in $PATH
          # There is no assertion that the file exists at build, use your own discretion  
          Exe = "lutris";
          Icon = "${pkgs.lutris}/share/icons/hicolor/128x128/apps/net.lutris.Lutris.png";
          Tags = ["Gaming" "Emulation"];
        }
      ];
    };
  };
}
```

</details>

## Finding Your Steam User ID

Your Steam user ID can be found in several ways:

1. **File system path**: Look in `~/.local/share/Steam/userdata/` or `~/.steam/steam/userdata/` - the numeric folder name is your user ID
2. **Steam profile URL**: Check the URL of your profile page in your browser or in your Steam client.
3. **Steam ID finder websites**: Get your SteamID3 from [findsteamid.com](https://findsteamid.com) or [steamid.io](https://steamid.io/)

## Manual Usage Example

<details>

<summary>If you prefer to manually manage the VDF file generation</summary>

```nix
{
  config,
  lib,
  pkgs,
  inputs,
  ...
}: let
  # Add the overlay to easily access json2steamshortcut
  nixpkgs.overlays = [ inputs.json2steamshortcut.overlays.default ];
  # Alternatively:
  # inherit (inputs.json2steamshortcut.packages.${stdenv.hostPlatform.system}) json2steamshortcut;
  
  # Declare our shortcuts as JSON, or load from directly from file
  json = builtins.toJSON [
    {
      AppName = "Heroic Games Launcher";
      Exe = "${lib.getExe pkgs.heroic}";
    }
    {
      AppName = "PrismLauncher";
      Exe = "prismlauncher";
    }
  ];

  # Create the vdf file
  vdf = pkgs.runCommand "shortcuts.vdf" {
    nativeBuildInputs = [pkgs.json2steamshortcut];
  } "echo '${json}' | json2steamshortcut > $out";
in {
  # use home-manager to place our shortcuts.vdf at the correct location, make sure to use correct Steam user ID
  home.file.".steam/steam/userdata/158842264/config/shortcuts.vdf" = {
    source = vdf;
    force = true;
  };
  # For MacOS, (untested):
  #home.file."Library/Application Support/Steam/userdata/158842264/config/shortcuts.vdf" = {
  #  source = vdf;
  #  force = true;
  #};
}
```

</details>

## Home Manager Shortcut Fields

The home-manager module supports the following fields for each shortcut:

- `AppName` (required): Name of the application as it appears in Steam
- `Exe` (required): Path to the executable, can also be any executable available in $PATH
- `StartDir` (optional): Starting directory for the application
- `LaunchOptions` (optional): Launch options for the application
- `Icon` (optional): Path to icon file
- `Tags` (optional): List of tags for the shortcut

## Flake Outputs

This flake provides:

- `packages.default`: The json2steamshortcut CLI as a package
- `overlays.default`: Overlay to add json2steamshortcut CLI to nixpkgs
- `homeModules.default`: Home Manager module for declarative Steam shortcuts management

## Known Limitations

- **Steam overwrites files**: Steam will overwrite the shortcuts.vdf file when you add software through the Steam interface.

  Software added through Steam will be lost with the next home-manager update.
  
  Use `services.steam-shortcuts.overwriteExisting = true` to implicitly overwrite any existing `shortcuts.vdf` (recommended.)

  This will only affect manually added non-steam games, and not your existing Steam game configurations such as launch options or compatibility tools.

## Potential Additions

These features might be out of the scope of this project, but PRs are welcome if someone wants to implement these functionalities.

- **Update shortcuts.vdf at runtime** Add merging support during runtime when shortcuts.vdf gets updated by Steam.
  Apply our declarative JSON to any existing shortcuts.vdf

- **Launch Steam Shortcuts from CLI** In theory, we could start our shortcuts using `steam steam://rungameid/<hash>`.
  Where we use Steams own hashing algorithm to generate a valid rungameid to start our non-steam games.
  [My](https://github.com/Joaqim) attempts to start these shortcuts have failed, here's a simple [typescript sandbox](https://kempo.io/projects/cmd3go6dx0008sea4v9v1lfvr) that _should_ be able to create valid shortcut hashes.

- **AppId**: For now we don't set AppId at all, but should be easy to implement if there was a need for it.
  As I haven't done any research on Steams TOS with using Steams own AppIds for non-steam game from a different store, I've left this untouched.

- **Icon**: Maybe dynamically find icon by executable name, or if icon is provide by icon identifier such as in a .desktop file:

  > `Icon = "net.lutris.Lutris.png"` -> "/usr/share/icons/hicolor/128x128/apps/net.lutris.Lutris.png"

- **Make Shortcut from package**

  Naive implementation:

  ```nix
  mkSteamShortcut = (myPkg: {
    Exe = lib.getExe myPkg;
    # TODO: Use logic to determine icon
    Icon = "${myPkg}/share/icons/hicolor/128x128/apps/*.png";
    # TODO: Does derivation always provide a nice readable name ? 
    AppName = myPkg.name; 
  })
  ```

  Usage:

  ```nix
    json = builtins.toJSON [
      (mkSteamShortcut pkgs.prismlauncher)
      (mkSteamShortcut pkgs.lutris)
    ];
    ...
  ```

  Alternatively create steam shortcut directly by parsing manually given desktop file:

  ```nix
  DesktopFile = "${pkgs.lutris}/share/applications/net.lutris.Lutris.desktop"
  ```
  