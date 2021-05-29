# Principle Investigator

The PR service is running by default on port `3002`.

Run:
```
cd service
go run .
```

Currently, the frontend does not support registration of a new study. So we do it manually with the following command:
```bash
curl -X POST -H "Content-Type: application/json" \
    -d '{"studyId": "1", "Users": ["patient1", "patient2"]}' \
    http://localhost:3002/api/register-study
```

