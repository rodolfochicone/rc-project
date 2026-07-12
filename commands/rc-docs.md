---
description: Generate or refresh RC project documentation — README, Postman collection, and OpenAPI spec.
disable-model-invocation: true
---

You are running the **RC documentation phase**. Run these skills in order, using the Skill tool:

1. **README** — invoke `rc-readme` to generate or update the project README from the built code.
2. **Postman** — invoke `rc-postman` to generate or update the Postman collection from the project's HTTP API.
3. **OpenAPI** — invoke `rc-openapi` to generate or update the OpenAPI spec from the project's endpoints.

If the project exposes no HTTP API, skip the Postman and OpenAPI steps and say so. Report the files written at the end.
