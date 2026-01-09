#!/bin/sh

set -eu

ROOT="${1:-.}"

# This script is intended to be used as a pre-commit hook.
# It ensures that every Go type implementing:
#   func (T) ObjectType() string
# is registered somewhere via:
#   astral.DefaultBlueprints.Add(&T{})

# Collect *.go files (exclude vendor and hidden dirs)
GO_FILES=$(find "$ROOT" \
  -type f -name "*.go" \
  ! -path "*/vendor/*" \
  ! -path "*/.*/*" \
  2>/dev/null || true)

if [ -z "$GO_FILES" ]; then
  echo "No Go files found under $ROOT"
  exit 0
fi

# Extract receiver type names that define ObjectType() string.
# Handles: (T), (t T), (*T), (t *T), (t pkg.T), (t *pkg.T)
# Skip if the line above has // astral:blueprint-ignore
OBJECT_TYPES=$(
  perl -0777 -ne '
    while (/(^.*?\n)?^[[:space:]]*func[[:space:]]*\([^)]*\)[[:space:]]*ObjectType[[:space:]]*\(\)[[:space:]]*string/gms) {
      my $prev = $1 || "";
      my $line = $&;
      next if $prev =~ /\/\/ astral:blueprint-ignore/;
      if ($line =~ /func\s*\(\s*([^)]*)\)\s*ObjectType/) {
        my $recv = $1;
        $recv =~ s/^[[:space:]]+|[[:space:]]+$//g;
        $recv =~ s/.*[[:space:]]+\*?//;
        $recv =~ s/^\*//;
        $recv =~ s/.*\.//;
        print "$recv\n" if $recv =~ /^[A-Za-z_][A-Za-z0-9_]*$/;
      }
    }
  ' $GO_FILES | sort -u
)

# Extract types registered via DefaultBlueprints.Add(&T{...}).
# Accept both `astral.DefaultBlueprints.Add(&T{})` and `DefaultBlueprints.Add(&T{})`.
# Also accept var t T; ... DefaultBlueprints.Add(&t)
# Includes multi-line Add calls by reading files as a stream and matching across newlines.
BLUEPRINT_TYPES=$(
  # First, from &Type{}
  perl -0777 -ne '
    while (/(?:^|\W)(?:[A-Za-z_][A-Za-z0-9_]*\.)?DefaultBlueprints\s*\.\s*Add\s*\(\s*&\s*([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)?)\s*(?:\{|\))/gms) {
      my $t = $1;
      $t =~ s/.*\.//; # drop package qualifier
      print "$t\n";
    }
  ' $GO_FILES \
  | sort -u
  # Second, from var t Type; ... Add(&t)
  perl -0777 -ne '
    while (/var\s+([A-Za-z_][A-Za-z0-9_]*)\s+([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)?)\s*.*?(?:[A-Za-z_][A-Za-z0-9_]*\.)?DefaultBlueprints\s*\.\s*Add\s*\(\s*&\s*\1\s*\)/gms) {
      my $t = $2;
      $t =~ s/.*\.//; # drop package qualifier
      print "$t\n";
    }
  ' $GO_FILES \
  | sort -u
)

FAILED=0

for t in $OBJECT_TYPES; do
  if ! echo "$BLUEPRINT_TYPES" | grep -qx "$t"; then
    echo "ERROR: Astral object type '$t' defines ObjectType() string but is not registered via DefaultBlueprints.Add(&${t}{...})"

    # Print full file path(s) where ObjectType() is defined for this type.
    # Use grep -nH to include filename; then filter by receiver type.
    echo "$GO_FILES" \
      | xargs grep -nH -E '^[[:space:]]*func[[:space:]]*\([^)]*\)[[:space:]]*ObjectType[[:space:]]*\(\)[[:space:]]*string' \
      | grep -E "\(\s*([A-Za-z_][A-Za-z0-9_]*\s+)?\*?([A-Za-z_][A-Za-z0-9_]*\.)?${t}\s*\)" \
      | sed -E 's/:([0-9]+):.*/:\1/' \
      | sed -E 's/^/  at /'

    FAILED=1
  fi
done

if [ "$FAILED" -ne 0 ]; then
  echo
  echo "Blueprint registration check FAILED."
  echo
  echo "Hint: add in an init() function:"
  echo "  _ = astral.DefaultBlueprints.Add(&<Type>{})"
  exit 1
fi

echo "Blueprint registration check passed."
