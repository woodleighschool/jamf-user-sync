# jamf-user-sync

Keeps managed Jamf Macs and iOS/iPadOS devices aligned with our Active Directory user records while Jamf remains deployed. AD owns the username, name, email, building and department fields; annual student year-level changes flow through the department mapping.

## ⚙️ Configuration

| Environment variable | Purpose                                                  | Default                       |
| -------------------- | -------------------------------------------------------- | ----------------------------- |
| `JAMF_HOST`          | Jamf hostname or HTTPS base URL                          | Required                      |
| `JAMF_CLIENT_ID`     | Jamf OAuth client ID                                     | Required                      |
| `JAMF_CLIENT_SECRET` | Jamf OAuth client secret                                 | Required                      |
| `LDAP_HOST`          | AD hostname or `ldap://` / `ldaps://` URL                | Required; bare hosts use LDAP |
| `LDAP_USERNAME`      | `DOMAIN\user` for NTLM, or a simple-bind principal       | Required                      |
| `LDAP_CREDENTIALS`   | Base64-encoded UTF-8 password                            | Required                      |
| `LDAP_BASE_DN`       | AD search base                                           | `dc=woodleighschool,dc=net`   |
| `DRY_RUN`            | Inspect changes without writing to Jamf                  | `false`                       |
| `RUN_TIMEOUT`        | Total run deadline                                       | `15m`                         |
| `REQUEST_TIMEOUT`    | Jamf requests, LDAP connection/bind and account searches | `30s`                         |
| `LOG_LEVEL`          | Diagnostic level (`debug`, `info`, `warn`, `error`)      | `info`                        |

The Jamf API role needs computer and mobile inventory read/update privileges. The directory account needs access to `sAMAccountName`, `name`, `mail`, `userAccountControl`, `Campus` and `department`.

## ▶️ Running

```sh
jamf-user-sync
jamf-user-sync --dry-run
jamf-user-sync --version
```

Each invocation validates configuration, fetches all managed-device inventory pages, reconciles user fields and exits. Kubernetes CronJobs in `wood-ops` own scheduling, retries and concurrency. `--dry-run` overrides `DRY_RUN`; `--dry-run=false` explicitly enables writes.

Building mappings are Senior Campus → `1`, Penbank → `2`, Minimbah → `3`. Department IDs retain the existing year-level and staff mappings. Unknown AD values map to `-1`. Accounts with the disabled `userAccountControl` bit (`0x2`) map to building `4` and department `26`. Devices without an assigned username or a matching AD account are skipped without changing their Jamf record. Ambiguous or malformed AD accounts fail that device.

Only differing user fields trigger an update, and unrelated inventory fields remain untouched. A JSON report on stdout counts devices, unassigned records, missing accounts, unchanged records, proposed/actual updates and lookup/update failures. Per-device failures on stderr include device kind, inventory ID and failed operation without user names, email addresses, credentials or API response bodies. Inventory/startup failures, cancellation and any failed device return a nonzero exit status; other devices continue after individual lookup or update failures.

## 🛠️ Development

```sh
mise install
mise run deps
mise run check
mise run container-check
```

Focused tests use synthetic directory entries, fakes and local SDK contract servers. They cover year-level/disabled-account mappings, field equality, dry-run, failure continuation, full pagination and narrow updates including empty name/email values. Normal checks do not contact Jamf or AD.

## 📦 Releases

Release Please maintains the existing version lineage. The shared container workflow publishes semver images at `ghcr.io/woodleighschool/jamf-user-sync` for `linux/amd64` and `linux/arm64`, with the organisation's signing, provenance and SBOM conventions. The non-root image runs the same one-shot command. `wood-ops` consumes version and digest pins through Renovate.

The command uses the maintained [Jamf Pro SDK](https://github.com/deploymenttheory/go-sdk-jamfpro-v2) for OAuth, transport, retries and pagination, and [go-ldap](https://github.com/go-ldap/ldap) for LDAP/NTLM. Computer inventory uses the v4 API, available in Jamf Pro 11.30 and later. SDK request builders send the five owned fields without serializing unrelated inventory sections.
