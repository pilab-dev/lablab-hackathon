import React, { useState } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import { Box, Skeleton, Alert, FormControl, InputLabel, Select, MenuItem, Typography } from '@mui/material';
import { useHistoryData } from '../hooks/useApi';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

interface PriceChartProps {
  symbol: string;
  timeframe?: '1h' | '4h' | '1d' | '1w';
  height?: number;
}

export const PriceChart: React.FC<PriceChartProps> = ({
  symbol,
  timeframe: initialTimeframe = '1d',
  height = 400
}) => {
  const [timeframe, setTimeframe] = useState<'1h' | '4h' | '1d' | '1w'>(initialTimeframe);
  const { data, isLoading, error } = useHistoryData(symbol, timeframe, 100);

  if (isLoading) {
    return (
      <Box sx={{ mt: 2 }}>
        <Skeleton variant="rectangular" height={30} sx={{ mb: 1 }} />
        <Skeleton variant="rectangular" height={height} />
      </Box>
    );
  }

  if (error || !data?.data || data.data.length === 0) {
    return (
      <Box sx={{ mt: 2 }}>
        <Alert severity="warning" sx={{ mb: 1 }}>
          No chart data available for {symbol} ({timeframe})
        </Alert>
      </Box>
    );
  }

  const chartData = {
    labels: data.data.map(point =>
      new Date(point.timestamp!).toLocaleString()
    ),
    datasets: [
      {
        label: 'Close Price',
        data: data.data.map(point => point.close),
        borderColor: 'rgb(75, 192, 192)',
        backgroundColor: 'rgba(75, 192, 192, 0.1)',
        tension: 0.1,
        fill: true,
        pointRadius: data.data.length > 50 ? 0 : 2,
        pointHoverRadius: 4,
      },
    ],
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
      tooltip: {
        mode: 'index' as const,
        intersect: false,
      },
    },
    scales: {
      y: {
        beginAtZero: false,
        ticks: {
          callback: function(value: number | string) {
            return typeof value === 'number' ? value.toFixed(2) : value;
          }
        }
      },
      x: {
        ticks: {
          maxTicksLimit: 8,
        }
      }
    },
    interaction: {
      mode: 'nearest' as const,
      axis: 'x' as const,
      intersect: false,
    },
  };

  return (
    <Box sx={{ mt: 2 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
        <Typography variant="subtitle2" color="text.secondary">
          {symbol}
        </Typography>
        <FormControl size="small" sx={{ minWidth: 80 }}>
          <InputLabel>TF</InputLabel>
          <Select
            value={timeframe}
            label="TF"
            onChange={(e) => setTimeframe(e.target.value as '1h' | '4h' | '1d' | '1w')}
          >
            <MenuItem value="1h">1H</MenuItem>
            <MenuItem value="4h">4H</MenuItem>
            <MenuItem value="1d">1D</MenuItem>
            <MenuItem value="1w">1W</MenuItem>
          </Select>
        </FormControl>
      </Box>
      <Box height={height}>
        <Line data={chartData} options={options} />
      </Box>
    </Box>
  );
};
