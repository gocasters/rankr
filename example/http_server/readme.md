# HTTP Server with OpenTelemetry Example

This guide explains how to run the example Go application, which is instrumented with OpenTelemetry to send traces to a Jaeger UI via the OpenTelemetry Collector.

### Prerequisites

* Go (1.18+)
* Docker & Docker Compose

### Running the Example

**1. Start Backend Services**

Run the OpenTelemetry Collector and Jaeger using Docker Compose. In your terminal, execute:

```bash
  docker-compose up -d
```

This will start the necessary services in the background.

**2. Run the HTTP Server**

In a new terminal window, run the Go application:

```bash
  go run main.go
```

You should see the message: `Server with Otel is ready.`

**3. Send a Test Request**

To generate a trace, send an HTTP request to the instrumented endpoint:

```bash
  curl http://localhost:8080/ping-otel
```

**4. View the Trace in Jaeger**

Open your web browser and navigate to the Jaeger UI:
[**http://localhost:16686**](http://localhost:16686)

In the UI:

* Select the `httpserver` service from the dropdown menu.
* Click the **"Find Traces"** button.
* You will see a trace for the `/ping-otel` request. Click on it to view the detailed waterfall diagram.

### Cleanup

To stop and remove the Docker containers after you are finished, run:

```bash
  docker-compose down -v
```