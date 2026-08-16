#!/usr/bin/env bash
#
# convert-image.sh
# Convert JPEG/HEIC/PNG images to optimized WebP for web usage.
#
# Usage:
#   ./scripts/convert-image.sh [options] <image1> [image2 ...]
#
# Examples:
#   ./scripts/convert-image.sh -t hero photo.heic
#   ./scripts/convert-image.sh -t carousel *.jpg
#   ./scripts/convert-image.sh -t hero -o web/static/img/ banner.heic
#

set -euo pipefail

# Check for ImageMagick
if command -v magick &>/dev/null; then
    MAGICK_CMD="magick"
elif command -v convert &>/dev/null; then
    MAGICK_CMD="convert"
else
    echo "Error: ImageMagick (magick or convert) is required but not found in PATH." >&2
    exit 1
fi

# Default configuration
TYPE="carousel"
OUTPUT_DIR=""
CUSTOM_QUALITY=""
CUSTOM_MAX_SIZE=""
CUSTOM_PREFIX=""
PREFIX_SET=false
FORCE=false

print_usage() {
    cat <<EOF
Usage: $(basename "$0") [options] <file1> [file2 ...]

Options:
  -t, --type <hero|carousel|content|general>
                        Image preset type (default: carousel)
                        - hero: max 1920px, quality 82, adds 'hero-' prefix
                        - carousel/content/general: max 1024px, quality 80
  -o, --output-dir <dir>
                        Target output directory (default: same as input file)
  -q, --quality <1-100> WebP compression quality (overrides preset default)
  -s, --max-size <px>   Maximum width/height in pixels (overrides preset default)
  -p, --prefix <str>    Custom filename prefix (set empty "" to omit prefix)
  -f, --force           Overwrite existing files without prompting
  -h, --help            Show this help message

Examples:
  $(basename "$0") -t hero IMG_1020.HEIC
  $(basename "$0") -t carousel photo1.jpg photo2.png
  $(basename "$0") -t hero -o web/static/img/ my-banner.heic
EOF
}

# Parse command line options
POSITIONAL_ARGS=()
while [[ $# -gt 0 ]]; do
    case "$1" in
        -t|--type)
            TYPE="$2"
            shift 2
            ;;
        --hero)
            TYPE="hero"
            shift
            ;;
        --carousel|--content|--general)
            TYPE="carousel"
            shift
            ;;
        -o|--output-dir)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -q|--quality)
            CUSTOM_QUALITY="$2"
            shift 2
            ;;
        -s|--max-size)
            CUSTOM_MAX_SIZE="$2"
            shift 2
            ;;
        -p|--prefix)
            CUSTOM_PREFIX="$2"
            PREFIX_SET=true
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        -*)
            echo "Error: Unknown option $1" >&2
            print_usage
            exit 1
            ;;
        *)
            POSITIONAL_ARGS+=("$1")
            shift
            ;;
    esac
done

if [[ ${#POSITIONAL_ARGS[@]} -eq 0 ]]; then
    echo "Error: No input files specified." >&2
    print_usage
    exit 1
fi

# Set preset parameters
case "$TYPE" in
    hero)
        MAX_SIZE=1920
        QUALITY=82
        DEFAULT_PREFIX="hero-"
        ;;
    carousel|content|general)
        MAX_SIZE=1024
        QUALITY=80
        DEFAULT_PREFIX=""
        ;;
    *)
        echo "Warning: Unknown type '$TYPE', defaulting to carousel preset." >&2
        MAX_SIZE=1024
        QUALITY=80
        DEFAULT_PREFIX=""
        ;;
esac

# Apply custom overrides if provided
if [[ -n "$CUSTOM_MAX_SIZE" ]]; then
    MAX_SIZE="$CUSTOM_MAX_SIZE"
fi
if [[ -n "$CUSTOM_QUALITY" ]]; then
    QUALITY="$CUSTOM_QUALITY"
fi
if [[ "$PREFIX_SET" = true ]]; then
    PREFIX="$CUSTOM_PREFIX"
else
    PREFIX="$DEFAULT_PREFIX"
fi

# Helper function to human-read format file sizes
format_size() {
    local bytes="$1"
    if [[ $bytes -ge 1048576 ]]; then
        awk -v b="$bytes" 'BEGIN { printf "%.2f MB", b / 1048576 }'
    elif [[ $bytes -ge 1024 ]]; then
        awk -v b="$bytes" 'BEGIN { printf "%.1f KB", b / 1024 }'
    else
        echo "${bytes} B"
    fi
}

echo "=================================================="
echo "Image Conversion to WebP"
echo "Preset: $TYPE (Max: ${MAX_SIZE}px, Quality: ${QUALITY}%, Prefix: '${PREFIX}')"
if [[ -n "$OUTPUT_DIR" ]]; then
    echo "Output Directory: $OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
fi
echo "=================================================="

TOTAL_ORIGINAL=0
TOTAL_NEW=0
SUCCESS_COUNT=0
FAIL_COUNT=0

for input_path in "${POSITIONAL_ARGS[@]}"; do
    if [[ ! -f "$input_path" ]]; then
        echo "⚠️  Skipping non-existent file: $input_path" >&2
        ((FAIL_COUNT++)) || true
        continue
    fi

    input_dir=$(dirname "$input_path")
    filename=$(basename "$input_path")
    base_name="${filename%.*}"

    # Avoid duplicating the prefix if it is already present
    if [[ -n "$PREFIX" && "$base_name" == "$PREFIX"* ]]; then
        out_base_name="$base_name"
    else
        out_base_name="${PREFIX}${base_name}"
    fi

    target_dir="${OUTPUT_DIR:-$input_dir}"
    output_path="${target_dir}/${out_base_name}.webp"

    if [[ -f "$output_path" && "$FORCE" = false && "$output_path" != "$input_path" ]]; then
        # Check if identical
        read -r -p "File '$output_path' already exists. Overwrite? [y/N] " confirm
        if [[ ! "$confirm" =~ ^[yY](es)?$ ]]; then
            echo "⏭️  Skipped: $filename"
            continue
        fi
    fi

    orig_size=$(stat -c%s "$input_path" 2>/dev/null || stat -f%z "$input_path" 2>/dev/null || echo 0)

    # Convert image with ImageMagick:
    # 1. -auto-orient: Fixes phone EXIF rotation
    # 2. -colorspace sRGB: Ensures accurate web colors
    # 3. -resize "${MAX_SIZE}x${MAX_SIZE}>": Downscale proportionally only if larger than MAX_SIZE
    # 4. -strip: Remove EXIF / GPS / device metadata
    # 5. -quality: WebP compression quality
    # 6. -define webp:method=6: Highest quality compression effort
    if "$MAGICK_CMD" "$input_path" \
        -auto-orient \
        -colorspace sRGB \
        -resize "${MAX_SIZE}x${MAX_SIZE}>" \
        -strip \
        -quality "$QUALITY" \
        -define webp:method=6 \
        "$output_path"; then

        new_size=$(stat -c%s "$output_path" 2>/dev/null || stat -f%z "$output_path" 2>/dev/null || echo 0)
        dimensions=$("$MAGICK_CMD" identify -format "%wx%h" "$output_path" 2>/dev/null || echo "unknown")

        TOTAL_ORIGINAL=$((TOTAL_ORIGINAL + orig_size))
        TOTAL_NEW=$((TOTAL_NEW + new_size))
        ((SUCCESS_COUNT++)) || true

        savings_pct=0
        if [[ $orig_size -gt 0 && $new_size -gt 0 ]]; then
            savings_pct=$(awk -v o="$orig_size" -v n="$new_size" 'BEGIN { printf "%.1f", ((o - n) / o) * 100 }')
        fi

        orig_formatted=$(format_size "$orig_size")
        new_formatted=$(format_size "$new_size")

        echo "✅ $filename -> $(basename "$output_path")"
        echo "   Dims: ${dimensions} | Size: ${orig_formatted} -> ${new_formatted} (${savings_pct}% saved)"
    else
        echo "❌ Failed to convert: $input_path" >&2
        ((FAIL_COUNT++)) || true
    fi
done

echo "--------------------------------------------------"
total_orig_formatted=$(format_size "$TOTAL_ORIGINAL")
total_new_formatted=$(format_size "$TOTAL_NEW")
total_savings_pct=0
if [[ $TOTAL_ORIGINAL -gt 0 && $TOTAL_NEW -gt 0 ]]; then
    total_savings_pct=$(awk -v o="$TOTAL_ORIGINAL" -v n="$TOTAL_NEW" 'BEGIN { printf "%.1f", ((o - n) / o) * 100 }')
fi

echo "Summary: $SUCCESS_COUNT converted, $FAIL_COUNT failed."
if [[ $SUCCESS_COUNT -gt 0 ]]; then
    echo "Total Size: ${total_orig_formatted} -> ${total_new_formatted} (${total_savings_pct}% total reduction)"
fi
