# Collector Framework

This repository contains the code for the blog post [Using the OpenTelemetry
Collector as a
framework](https://dmathieu.com/en/development/opentelemetry-collector-framework/).

## Running locally

To run the collector locally, you can use the `make run` command.
All components will be started, with the following entry and exit points:

* `localhost:2504` - The traces receiver endpoint.
* `localhost:1404` - The metrics exporter endpoint.
* `localhost:2806` - The extension healcheck endpoint.

If nothing listens on the exporter endpoint, errors will be raised regularly,
as the exporter will fail making outgoing HTTP requests.

## What components are there?

The intent is to demo building each kind of component for a Collector. So each
component kind at this time is exposed.

* **Receiver** - The receiver listens on `localhost:2504`, and emits a new trace for each receiver HTTP call (no matter the content of the request).
* **Processor** - The processor adds dummy attributes to the span.
* **Connector** - Transforms the trace into a metric.
* **Exporter** - The exporter makes an HTTP request to `localhost:1304` with the metric content turned into JSON.
* **Extension** - The extension creates an HTTP server in `localhost:2806` that just emits an healthcheck (without actually checking anything).
* **Scraper** - The scraper runs every 10 seconds and emits a metric supposed to be CPU usage (but the value will always be 45.5). The metric is emitted on the same pipeline as the one received from the connector.

## Don't do this at home.

These components are built purely for the purpose of demoing how you can extend
the OpenTelemetry Collector with custom components.

They should not be used in production, ever.
