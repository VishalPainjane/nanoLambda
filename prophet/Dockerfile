FROM python:3.10-slim

WORKDIR /app

# Install system dependencies for Prophet (if needed, though wheels usually exist)
# RUN apt-get update && apt-get install -y ...

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "-u", "main.py"]
