# gh-metrics

## Try it Out

Initial Execution

```bash
# example repo at github.com/repo-owner/my-repo
go run main.go -repo my-repo -owner repo-owner 
```

This will provide a summary report and store the data in `./metrics/<repo>_metrics_<timestamp>.json`


## Known Issues

Does not currently work - ""{\"message\":\"Pagination with the page parameter is not supported for large datasets, please use cursor based pagination (after/before)\",\"documentation_url\":\"https://docs.github.com/rest/issues/issues#list-repository-issues\",\"status\":\"422\"}"" 

Need to investigate

## Ideas

- Support for comparing files
- CI to run and commit metrics on cron