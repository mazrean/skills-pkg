# Go Module Integration

skills-pkg can fetch skills from **Go modules** using the Go Module proxy protocol. This lets you distribute skills as ordinary Go packages and take advantage of the existing module ecosystem — versioning, the module proxy cache, and `go.mod` lock-in.

## How it works

When `source = "go-mod"` is set for a skill, skills-pkg:

1. Resolves the version to install (see [Version resolution](#version-resolution))
2. Downloads the module zip from the configured proxy (or directly from the VCS)
3. Strips the standard module zip prefix (`{module}@{version}/`) and extracts the files
4. Copies the subdirectory specified by `subdir` into every `install_target`

## Version resolution

Version resolution follows this priority order when `--version` is not specified (or is empty):

1. **`go.mod` lookup** — skills-pkg walks up the directory tree from the current working directory to find the nearest `go.mod` file. If the module path appears in a `require` directive, that version is used. This keeps skills in sync with your Go dependency graph automatically.
2. **Latest from proxy** — if the module is not listed in any `go.mod`, skills-pkg queries `{proxy}/{module}/@latest` and uses the returned version.

Specifying `--version latest` explicitly skips the `go.mod` lookup and always fetches the latest version from the proxy.

```sh
# Uses version from go.mod if present, otherwise latest
skills-pkg add my-skill --source go-mod --url github.com/example/go-skills

# Always fetches the latest, ignoring go.mod
skills-pkg add my-skill --source go-mod --url github.com/example/go-skills --version latest

# Pin to a specific version
skills-pkg add my-skill --source go-mod --url github.com/example/go-skills --version v1.3.0
```

## GOPROXY support

skills-pkg respects the standard `GOPROXY` environment variable.

**Default:** `https://proxy.golang.org,direct`

The variable follows the same syntax as the Go toolchain:

| Syntax | Meaning |
|---|---|
| `https://proxy.example.com` | Use this proxy |
| `direct` | Fetch directly from the VCS over HTTPS |
| `off` | Disallow all downloads (error) |
| `A,B` | Try A; on failure fall back to B |
| `A\|B` | Try A; if A returns a non-404 error, also try B |

```sh
# Use a private proxy, fall back to the public proxy
GOPROXY=https://goproxy.mycompany.com,https://proxy.golang.org,direct skills-pkg install

# Fetch everything directly from VCS (no proxy)
GOPROXY=direct skills-pkg install
```

### `direct` mode behavior

When a proxy entry is `direct`, skills-pkg fetches the module by cloning the repository over HTTPS (`https://{module-path}`) using the embedded go-git library — no external `git` binary is required. It checks out the specified version as a tag first, then as a branch if the tag is not found.

## Environment variables

| Variable | Description |
|---|---|
| `GOPROXY` | Proxy list in Go toolchain format. Defaults to `https://proxy.golang.org,direct` |
| `SKILLSPKG_TEMP_DIR` | Override the base directory used for temporary module downloads. Defaults to the OS temp directory |

## Packaging skills as a Go module

Any Go module can serve as a skill source. The recommended layout is:

```
your-module/
├── go.mod
├── skills/
│   ├── skill-a/        ← subdir = "skills/skill-a"
│   │   ├── SKILL.md
│   │   └── ...
│   └── skill-b/        ← subdir = "skills/skill-b"
│       ├── SKILL.md
│       └── ...
└── (other Go source files)
```

The default `subdir` when using `add` is `skills/<name>`, which matches this layout.

Publish the module with a semver tag (e.g., `git tag v1.0.0 && git push origin v1.0.0`). Once pushed, the module becomes available through the Go module proxy within minutes.

```sh
# Install from your published module
skills-pkg add skill-a --source go-mod --url github.com/yourorg/your-module
```
