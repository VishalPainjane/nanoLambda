# nanolambda

> the intelligent serverless platform
> aws lambda meets facebook prophet

nanolambda is a research-grade serverless platform that uses ai to predict traffic spikes and pre-warm containers, eliminating cold starts.

## features

- **scale-to-zero:** containers automatically stop when idle to save resources.
- **ai prediction:** uses time-series forecasting (prophet) to predict when users will arrive.
- **pre-warming:** starts containers *before* traffic hits, ensuring instant responses.
- **developer experience:** polished cli (`nanolambda`) for deploying and managing functions.