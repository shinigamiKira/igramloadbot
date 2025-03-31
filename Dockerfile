# Use official Python image
FROM python:3.9-slim

# Set working directory
WORKDIR /app

# Copy requirements first to leverage Docker cache
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application files
COPY . .

# Environment variables for secrets
ENV BOT_TOKEN=${BOT_TOKEN}
ENV CHAT_ID=${CHAT_ID}

# Create downloads directory
RUN mkdir -p downloads

# Run the application
CMD ["python", "scraper.py"]
