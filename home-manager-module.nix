{
  config,
  lib,
  pkgs,
  ...
}: let
  cfg = config.services.steam-shortcuts;

  # This hardcoded steam path seems pretty consistent, at least for Linux
  steamConfDirRelative =
    if pkgs.stdenv.hostPlatform.isDarwin
    then "Library/Application Support/Steam"
    else ".steam/steam";
in
  with lib; {
    options.services.steam-shortcuts = {
      enable = mkEnableOption "Steam shortcuts management";

      package = mkOption {
        type = types.package;
        default = pkgs.callPackage ./default.nix {};
        defaultText = literalExpression "inputs.json2steamshortcut.packages.\${stdenv.hostPlatform.system}.json2steamshortcut";
      };
      steamUserId = mkOption {
        type = types.int;
        description = ''
          The Steam user ID is the numeric identifier for your Steam account.
          It can be found in the Steam userdata directory, typically located at
          ~/.local/share/Steam/userdata/<steamUserId>/
          or
          ~/.steam/steam/userdata/<steamUserId>/

          You can also find your Steam user ID by checking the URL of your profile page
          in your browser or in your Steam client.
        '';

        example = 158842264;
      };

      # If provided, this will assume that steam is run with $HOME set to this directory
      # Only useful in very particular cases where one has already configured to run Steam
      # with $HOME set to to a non-standard location, e.g. /mnt/Steam
      steamHomeDir = mkOption {
        type = types.str;
        default = config.home.homeDirectory;
        description = "Steam home directory, only change if you know what you are doing";
      };

      userConfigDir = mkOption {
        type = types.str;
        internal = true;
        default = "${cfg.steamHomeDir}/${steamConfDirRelative}/userdata/${builtins.toString cfg.steamUserId}/config";
      };

      # NOTE: This configuration is formatted for use with json2steamshortcut
      #       and will be converted to JSON and piped via stdout to json2steamshortcut.
      #
      #       Notably, Steam uses lowercase keys for most of these options,
      #       excepting StartDir and LaunchOptions
      shortcuts = mkOption {
        type = types.listOf (types.submodule {
          options = {
            AppName = mkOption {
              type = types.str;
              description = "Name of the application as it appears in Steam";
            };
            Exe = mkOption {
              type = types.str;
              description = "Path to the executable";
            };
            AppId = mkOption {
              type = with types; nullOr str;
              default = null;
              internal = true; # Disabled for now
              description = "Steam AppID";
            };
            StartDir = mkOption {
              type = with types; nullOr str;
              default = null;
              description = "Starting directory for the application";
            };
            LaunchOptions = mkOption {
              type = with types; nullOr str;
              default = null;
              description = "Launch options for the application";
            };
            Icon = mkOption {
              type = with types; nullOr str;
              default = null;
              description = "Path to icon file";
            };
            Tags = mkOption {
              type = with types; nullOr (listOf str);
              default = null;
              description = "Tags for the shortcut";
            };
          };
        });
        default = [];
        description = "List of shortcuts to create";
        example = literalExpression ''
          [
            {
              AppName = "Performous";
              Exe = "''${pkgs.performous}/bin/performous";
            }
            {
              AppName = "Firefox";
              Exe = "''${pkgs.firefox}/bin/firefox";
              LaunchOptions = "--new-window";
            }
          ];
        '';
      };
    };

    config = mkIf cfg.enable {
      assertions = [
        {
          assertion = cfg.shortcuts != [];
          message = "services.steam-shortcuts expects at least one shortcut to be defined";
        }
        {
          assertion = cfg.steamHomeDir != "";
          message = "Steam Home directory could not be determined";
        }
      ];

      # Check for shortcuts with duplicate AppNames
      # Since AppName + Exe is used by Steam to generate
      # non-steam AppId for: `steam steam://rungameid/<hash>`
      # NOTE: We are intentionally not including 'Exe' in search key
      # since we are prohibited from using store paths as strings
      warnings = let
        firstDuplicate =
          (builtins.foldl'
            (acc: value:
              if acc.found != null
              then acc
              else if acc.seen ? ${value}
              then {
                inherit (acc) seen;
                found = value;
              }
              else {
                seen = acc.seen // {${value} = true;};
                found = null;
              })
            {
              seen = {};
              found = null;
            }
            (builtins.map (attr: attr.AppName) cfg.shortcuts)).found;
      in
        lib.optional (firstDuplicate != null) "services.steam-shortcuts: Found duplicate AppName: ${firstDuplicate} - this may cause issues when creating shortcuts from Steam to launch non-steam applications";

      # Create shortcuts.vdf file
      home.file."${cfg.userConfigDir}/shortcuts.vdf" = let
        # Utility to filter out shortcut fields with null values
        cleanAttrs = attrs:
          lib.attrsets.filterAttrs (_key: value: value != null) attrs;

        shortcuts = lib.map cleanAttrs cfg.shortcuts;

        json = builtins.toJSON shortcuts;
        vdf = pkgs.runCommandLocal "shortcuts.vdf" {
          nativeBuildInputs = [cfg.package];
        } "echo '${json}' | json2steamshortcut > $out";
      in {
        source = vdf;
        force = true;
      };
    };
  }
