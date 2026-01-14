FROM python:3.11-slim

# Install system dependencies if needed
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install python dependencies for the runner
RUN pip install flask

# Create directories
RUN mkdir -p /runner /function

# Copy the runner code
COPY runner.py /runner/runner.py

# Set working directory to where user code lives
WORKDIR /function

# Expose port
EXPOSE 8080

# Default command
CMD ["python", "/runner/runner.py"]
