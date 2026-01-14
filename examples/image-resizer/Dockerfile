FROM nanolambda/base-python:3.11

# Install system dependencies for Pillow (image library)
# (base-python is slim, so might need zlib/jpeg libs if not present, but usually okay for basic)

COPY handler.py /function/handler.py
COPY requirements.txt /function/requirements.txt

# Install the dependencies
RUN pip install --no-cache-dir -r /function/requirements.txt
