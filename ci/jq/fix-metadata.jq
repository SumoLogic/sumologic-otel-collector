def fix_cache_manifest:
  .
  | select(has("cache.manifest"))
  | .["cache.manifest"] = (.["cache.manifest"] | fromjson);

def fix_metadata:
  .
  | map_values(fix_cache_manifest);

fix_metadata
