import requests
import json

class WarmupClient:
    def __init__(self, gateway_url="http://localhost:8080"):
        self.url = gateway_url

    def trigger_warmup(self, function_name):
        try:
            url = f"{self.url}/admin/warmup"
            payload = {"function": function_name}
            headers = {'Content-Type': 'application/json'}
            
            response = requests.post(url, json=payload, headers=headers, timeout=2)
            
            if response.status_code == 200:
                print(f"[Warmup] Successfully triggered for {function_name}")
                return True
            else:
                print(f"[Warmup] Failed for {function_name}: Status {response.status_code}")
                return False
                
        except requests.exceptions.RequestException as e:
            print(f"[Warmup] Connection error for {function_name}: {e}")
            return False
