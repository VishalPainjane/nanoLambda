import os
import sys
import importlib.util
from flask import Flask, request, jsonify
import traceback

app = Flask(__name__)

# Global variable to hold the user's function module
user_module = None

def load_user_function():
    global user_module
    function_path = "/function/handler.py"
    
    if not os.path.exists(function_path):
        print(f"Error: Function file not found at {function_path}")
        return False

    try:
        spec = importlib.util.spec_from_file_location("handler", function_path)
        user_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(user_module)
        print(f"Successfully loaded function from {function_path}")
        return True
    except Exception as e:
        print(f"Error loading function: {e}")
        traceback.print_exc()
        return False

@app.route('/invoke', methods=['POST'])
def invoke():
    if user_module is None:
        return jsonify({"error": "Function not loaded"}), 500

    try:
        # Get JSON body or default to empty dict
        req_data = request.get_json(force=True, silent=True) or {}
        
        # Call the user's 'handle' function
        if hasattr(user_module, 'handle'):
            result = user_module.handle(req_data)
            return jsonify(result)
        else:
            return jsonify({"error": "Function 'handle' not found in handler.py"}), 500

    except Exception as e:
        traceback.print_exc()
        return jsonify({"error": str(e)}), 500

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "ready"}), 200

if __name__ == '__main__':
    # Load function on startup
    success = load_user_function()
    if not success:
        print("WARNING: Failed to load user function on startup")
    
    # Run on port 8080
    app.run(host='0.0.0.0', port=8080)
