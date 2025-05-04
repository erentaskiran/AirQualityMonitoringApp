"use client";

import { useEffect, useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { MakeRequest } from '@/lib/utils';

interface ChartDataPoint {
  timestamp: number; 
  timeLabel: string; 
  value: number;
  parameter: string;
}

export default function ChartsPage() {
  const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    
    const endTime = new Date().toISOString();
    const startTime = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString();
    
    const fetchAnomalies = async () => {
      try {
        const response = await MakeRequest(
          'api/anomalies/timerange', 
          "GET", 
          null, 
          {
            "X-Start-Time": startTime,
            "X-End-Time": endTime
          }
        );
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const receivedData = await response.json();
        
        if (Array.isArray(receivedData)) {
          const processedData = receivedData
            .map(item => {
              const date = new Date(item.time);
              return {
                timestamp: date.getTime(), 
                timeLabel: date.toLocaleString(), 
                value: item.value,
                parameter: item.parameter,
              };
            })
            .filter(item => !isNaN(item.timestamp) && typeof item.value === 'number') 
            .sort((a, b) => a.timestamp - b.timestamp); 

          setChartData(processedData);
          console.log("Processed chart data:", processedData);
        } else {
          console.warn("Received non-array data from API:", receivedData);
          setError("Received unexpected data format.");
        }
      } catch (err) {
        console.error("Error fetching anomalies:", err);
        setError(`Failed to fetch anomaly data: ${err instanceof Error ? err.message : 'Unknown error'}`);
      } finally {
        setLoading(false);
      }
    };
    
    fetchAnomalies();
    
    const intervalId = setInterval(fetchAnomalies, 30000); 
    
    return () => {
      clearInterval(intervalId);
    };
  }, []); 

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', width: '100%' }}> 
      <h1>Air Quality Charts</h1>
      <p>Showing recent air quality anomaly values over time.</p>

      {loading && <p>Loading chart data...</p>}
      {error && <p style={{ color: 'red' }}>Error: {error}</p>}

      {!loading && !error && chartData.length === 0 && (
        <p>No anomaly data available to display.</p>
      )}

      {!loading && !error && chartData.length > 0 && (
        <div style={{ width: '100%', height: 400 }}> 
          <ResponsiveContainer>
            <LineChart
              data={chartData}
              margin={{
                top: 5,
                right: 30,
                left: 20,
                bottom: 5,
              }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                 dataKey="timestamp"
                 type="number" 
                 scale="time" 
                 domain={['dataMin', 'dataMax']} 
                 tickFormatter={(unixTime) => new Date(unixTime).toLocaleTimeString()} 
                 name="Time"
               />
              <YAxis label={{ value: 'Value', angle: -90, position: 'insideLeft' }} />
              <Tooltip labelFormatter={(label) => new Date(label).toLocaleString()} />
              <Legend />
              <Line
                 type="monotone"
                 dataKey="value"
                 stroke="#8884d8"
                 activeDot={{ r: 8 }}
                 name="Anomaly Value"
                 connectNulls 
               />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}