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
import { useAssets } from '../hooks/useApi';

export function AssetsPage() {
  const { data, isLoading } = useAssets(true);

  if (isLoading) return <CircularProgress />;

  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 3 }}>Assets</Typography>
      <TableContainer component={Card}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Symbol</TableCell>
              <TableCell>Class</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Decimals</TableCell>
              <TableCell>Collateral Value</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.assets?.map((asset) => (
              <TableRow key={asset.altname} hover>
                <TableCell>
                  <Typography fontWeight="bold">{asset.altname}</Typography>
                </TableCell>
                <TableCell>{asset.aclass}</TableCell>
                <TableCell>
                  <Chip
                    label={asset.status}
                    color={asset.status === 'enabled' ? 'success' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell>{asset.decimals}</TableCell>
                <TableCell>{asset.collateral_value?.toFixed(4) || '-'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <Typography variant="caption" color="text.secondary" sx={{ mt: 1 }}>
        {data?.count} assets total
      </Typography>
    </Box>
  );
}
