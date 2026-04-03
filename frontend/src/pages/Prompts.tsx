import {
  Box,
  Card,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  CircularProgress,
} from '@mui/material';
import { usePrompts } from '../hooks/useApi';

export function PromptsPage() {
  const { data, isLoading } = usePrompts(50);

  if (isLoading) return <CircularProgress />;

  const getActionColor = (action?: string) => {
    switch (action?.toLowerCase()) {
      case 'buy': return 'success';
      case 'sell': return 'error';
      default: return 'default';
    }
  };

  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 3 }}>Prompt History</Typography>
      <TableContainer component={Card}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Time</TableCell>
              <TableCell>Pair</TableCell>
              <TableCell>Action</TableCell>
              <TableCell>Size</TableCell>
              <TableCell>Confidence</TableCell>
              <TableCell>Result</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.prompts?.map((prompt) => (
              <TableRow key={prompt.id} hover>
                <TableCell>
                  {prompt.created_at && new Date(prompt.created_at).toLocaleString()}
                </TableCell>
                <TableCell>{prompt.pair}</TableCell>
                <TableCell>
                  <Chip
                    label={prompt.action || 'analyze'}
                    color={getActionColor(prompt.action)}
                    size="small"
                  />
                </TableCell>
                <TableCell>{prompt.size_pct ? `${(prompt.size_pct * 100).toFixed(1)}%` : '-'}</TableCell>
                <TableCell>{prompt.confidence ? `${(prompt.confidence * 100).toFixed(0)}%` : '-'}</TableCell>
                <TableCell>
                  <Chip
                    label={prompt.success ? 'success' : 'pending'}
                    color={prompt.success ? 'success' : 'warning'}
                    size="small"
                  />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
