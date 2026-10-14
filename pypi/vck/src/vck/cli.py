import sys
from pathlib import Path

# Try to import from installed vivechak package, or local sibling in repo
try:
    from vivechak.cli import main
except ImportError:
    sibling_src = Path(__file__).resolve().parent.parent.parent.parent / "vivechak" / "src"
    if sibling_src.exists():
        sys.path.insert(0, str(sibling_src))
        from vivechak.cli import main
    else:
        raise

if __name__ == "__main__":
    main()
