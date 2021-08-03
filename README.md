# AWS service authorization reference

This is a JSON-formatted scrape of the [AWS Service Authorization Reference](https://docs.aws.amazon.com/service-authorization/latest/reference/reference.html), along with a Golang program to update it.

## NPM package

If you're using the NPM package, you can use the service reference like this:

```typescript
import { serviceAuth } from '@fluggo/aws-service-auth-reference';

for(const service of serviceAuth) {
  console.log(service.name);
}
```
