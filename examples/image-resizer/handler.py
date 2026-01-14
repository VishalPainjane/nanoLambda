from PIL import Image
import requests
import io
import base64

def handle(request):
    # 1. Get Input
    image_url = request.get('url', 'https://picsum.photos/1200/1200') # Random reliable image
    target_width = request.get('width', 100)
    target_height = request.get('height', 100)

    print(f"Downloading image from {image_url}...")

    # 2. Download Image
    try:
        response = requests.get(image_url, timeout=5)
        response.raise_for_status()
    except Exception as e:
        return {"error": f"Failed to download image: {str(e)}"}

    original_size = len(response.content)
    
    # 3. Process Image (The CPU intensive part)
    try:
        image = Image.open(io.BytesIO(response.content))
        original_dims = image.size
        
        # Resize
        image.thumbnail((target_width, target_height))
        
        # Save to buffer
        output_buffer = io.BytesIO()
        image.save(output_buffer, format=image.format or 'JPEG')
        compressed_data = output_buffer.getvalue()
        compressed_size = len(compressed_data)
        
        # Calculate savings
        savings_percent = ((original_size - compressed_size) / original_size) * 100
        
        return {
            "status": "success",
            "original_dimensions": f"{original_dims[0]}x{original_dims[1]}",
            "new_dimensions": f"{image.size[0]}x{image.size[1]}",
            "original_size_kb": f"{original_size / 1024:.2f} KB",
            "new_size_kb": f"{compressed_size / 1024:.2f} KB",
            "space_saved": f"{savings_percent:.1f}%",
            "message": "Image resized successfully! In a real app, this would now be saved to S3."
        }

    except Exception as e:
        return {"error": f"Failed to process image: {str(e)}"}
