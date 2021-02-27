# DynamoDB schema

## Repository

### Repository config
- PK: `repository_config:{registry}/{repository}`
- SK: `{timestamp}`
- PK: `list:repository`
- SK: `{registry}/{repository}`

### Image layer digests index
- PK: `layer_digest:{layer_digest}`
- SK: `{registry}/{repository}:{digest}`

## Vulnerability

### Vulnerability Info
- PK: `vuln_info:{vuln_id}`
- SK: `{pkg_type}/{pkg_name}`

### Scan report
- PK: `report:{registry}/{repository}:{tag}`
- SK: `{scan_type}/{timestamp}/{report_id}`
- PK: `list:report`
- SK: `{report_id}`

### Repository vulnerability status
- PK: `repo_vuln_status:{registry}/{repository}:{tag}`
- SK: `{vuln_id}:{pkg_source}:{pkg_name}:{updated_at}`
- PK2: `repo_vuln_status:{vuln_id}`
- SK2: `{registry}/{repository}:{tag}:{pkg_source}:{pkg_name}:{updated_at}`

## Team / member

### Team
- PK: `list:team`
- SK: `{team_id}`

### Team/Member map
- PK: `team_member_map:{team_id}`
- SK: `{member_id}`
- PK2: `team_member_map:{member_id}`
- SK2: `{team_id}`

### Team/Repository map
- PK: `team_repo_map:{team_id}`
- SK: `{registry}/{repository}`
- PK2: `team_repo_map:{registry}/{repository}`
- SK2: `{team_id}`
