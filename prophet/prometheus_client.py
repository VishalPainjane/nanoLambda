from prometheus_api_client import PrometheusConnect, MetricSnapshotDataFrame
from datetime import timedelta
import pandas as pd

class PrometheusClient:
    def __init__(self, url="http://localhost:9090"):
        self.prom = PrometheusConnect(url=url, disable_ssl=True)

    def get_function_data(self, function_name, lookback_days=7):
        """
        Fetches the request rate for a function over the last lookback_days.
        Returns a DataFrame suitable for Prophet (ds, y).
        """
        # Query: rate(http_requests_total{function="name"}[5m])
        # This gives requests per second. We multiply by 300 to get requests per 5 min block.
        query = f'rate(http_requests_total{{function="{function_name}"}}[5m]) * 300'
        
        start_time = timedelta(days=lookback_days)
        
        # Get metric data
        metric_data = self.prom.get_metric_range_data(
            metric_name=query,  # It accepts query here too
            start_time=start_time,
            end_time=timedelta(minutes=0), # Now
            chunk_size=timedelta(days=1),
        )
        
        if not metric_data:
            return None

        # Convert to DataFrame
        df = MetricSnapshotDataFrame(metric_data)
        
        # Rename columns for Prophet
        # Dataframe usually has: timestamp, value, ...
        # We need 'ds' and 'y'
        df['ds'] = pd.to_datetime(df['timestamp'], unit='s')
        df['y'] = pd.to_numeric(df['value'])
        
        return df[['ds', 'y']]
