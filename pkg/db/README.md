# DynamoDB schema

## StatusSequence

- PK: `seq:status`
- SK: `-`
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
- PK2: `list:vuln_info`
- SK2: `{created_at}/{vuln_id}`

### Scan report
- PK: `report:{registry}/{repository}:{tag}`
- SK: `{scan_type}/{timestamp}/{report_id}`
- PK2: `list:report`
- SK2: `{report_id}`

### Repository vulnerability status
- PK: `repo_vuln_status:{registry}/{repository}:{tag}`
- SK:
    - Package: `{vuln_id}:pkg:{pkg_source}:{pkg_name}`
- PK2: `repo_vuln_status:{vuln_id}`
- SK2:
    - Package: `{registry}/{repository}:{tag}:pkg:{pkg_source}:{pkg_name}`

### Repository vulnerability change log
- PK: `repo_vuln_changelog:{registry}/{repository}:{tag}`
- SK:
    - Package: `{vuln_id}:pkg:{pkg_source}:{pkg_name}:{seq}`
- PK2: `repo_vuln_changelog:{vuln_id}`
- SK2:
    - Package: `{registry}/{repository}:{tag}:pkg:{pkg_source}:{pkg_name}:{seq}`

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
