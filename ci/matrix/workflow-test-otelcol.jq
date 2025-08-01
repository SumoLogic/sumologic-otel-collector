def add_pkgs($include; $pkgs):
  [$include[] as $i | $pkgs[] as $p | $i + {pkg: $p}];

def generate_matrix($base; $pkgs):
  $base + {
    include: add_pkgs($base.include; $pkgs)
  };

generate_matrix($base; $pkgs)
