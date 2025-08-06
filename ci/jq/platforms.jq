def show_platform:
    if has("os.version") then
       "\(.os)(\(.["os.version"]))/\(.architecture)"
    else
       "\(.os)/\(.architecture)"
    end;

def show_platforms:
  .manifests
  | map(.platform | show_platform);

show_platforms
