import time
import schedule
from prophet import Prophet
import pandas as pd
import logging
from prometheus_client import PrometheusClient
from warmup import WarmupClient

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [Prophet] - %(message)s',
    datefmt='%H:%M:%S'
)
logger = logging.getLogger(__name__)

# Constants
PROMETHEUS_URL = "http://localhost:9090"
GATEWAY_URL = "http://localhost:8080"
PREDICTION_THRESHOLD = 5.0 # Requests expected in next window
CONFIDENCE_THRESHOLD = 0.75

def job():
    logger.info("Starting prediction cycle...")
    
    prom = PrometheusClient(url=PROMETHEUS_URL)
    warmup = WarmupClient(gateway_url=GATEWAY_URL)
    
    # 1. Discover functions from Prometheus
    # We query for all time series of http_requests_total to find unique function names
    try:
        # This is a bit hacky, but works to discover functions that have traffic
        # In a real system, we might query the Registry API instead
        metric_families = prom.prom.custom_query(query='count(http_requests_total) by (function)')
        functions = [m['metric']['function'] for m in metric_families]
    except Exception as e:
        logger.error(f"Failed to discover functions: {e}")
        return

    logger.info(f"Found functions: {functions}")

    for func_name in functions:
        try:
            # 2. Fetch History
            df = prom.get_function_data(func_name)
            
            if df is None or len(df) < 20: # Need enough data points
                logger.info(f"Not enough data for {func_name} (rows: {len(df) if df is not None else 0})")
                continue
                
            # 3. Train Model
            # suppress prophet logging
            model = Prophet(yearly_seasonality=False, weekly_seasonality=True, daily_seasonality=True)
            model.fit(df)
            
            # 4. Predict
            future = model.make_future_dataframe(periods=2, freq='5min') # Predict next 10 mins
            forecast = model.predict(future)
            
            # Get the prediction for the immediate future (last row)
            next_point = forecast.iloc[-1]
            predicted_y = next_point['yhat']
            # Prophet doesn't give direct "confidence" probability like a classifier, 
            # but we can look at yhat_lower/upper or just use yhat for now.
            # For the demo, we'll assume if yhat > threshold, we warmup.
            
            logger.info(f"Prediction for {func_name}: {predicted_y:.2f} requests (Threshold: {PREDICTION_THRESHOLD})")
            
            if predicted_y > PREDICTION_THRESHOLD:
                logger.info(f"TRIGGERING WARMUP for {func_name}")
                warmup.trigger_warmup(func_name)
            else:
                logger.info(f"No warmup needed for {func_name}")
                
        except Exception as e:
            logger.error(f"Error processing {func_name}: {e}")

if __name__ == "__main__":
    logger.info("NanoLambda Intelligence Layer (Prophet) Started")
    
    # Run immediately on startup
    job()
    
    # Schedule every 5 minutes
    schedule.every(5).minutes.do(job)
    
    while True:
        schedule.run_pending()
        time.sleep(1)
