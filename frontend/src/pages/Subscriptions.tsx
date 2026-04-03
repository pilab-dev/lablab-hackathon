import { useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  CircularProgress,
  Alert,
} from '@mui/material';
import { Add as AddIcon, Delete as DeleteIcon } from '@mui/icons-material';
import {
  useSubscriptionDetails,
  useAddSubscription,
  useRemoveSubscription,
} from '../hooks/useApi';

export function SubscriptionsPage() {
  const { data, isLoading, error } = useSubscriptionDetails();
  const addMutation = useAddSubscription();
  const removeMutation = useRemoveSubscription();
  const [open, setOpen] = useState(false);
  const [symbol, setSymbol] = useState('');

  const handleAdd = () => {
    if (symbol.trim()) {
      addMutation.mutate(symbol.trim(), {
        onSuccess: () => {
          setOpen(false);
          setSymbol('');
        },
      });
    }
  };

  const handleRemove = (sym: string) => {
    if (confirm(`Remove subscription for ${sym}?`)) {
      removeMutation.mutate(sym);
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5">Subscriptions</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setOpen(true)}
        >
          Add Subscription
        </Button>
      </Box>

      {isLoading ? (
        <CircularProgress />
      ) : error ? (
        <Alert severity="error">Failed to load subscriptions</Alert>
      ) : data?.subscriptions?.length === 0 ? (
        <Alert severity="info">No active subscriptions</Alert>
      ) : (
        <Box sx={{ display: 'grid', gap: 2 }}>
          {data?.subscriptions?.map((sub) => (
            <Card key={sub.symbol}>
              <CardContent sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Box>
                  <Typography variant="h6">{sub.symbol}</Typography>
                  <Box sx={{ display: 'flex', gap: 1, mt: 1 }}>
                    <Chip
                      label={sub.is_active ? 'Active' : 'Inactive'}
                      color={sub.is_active ? 'success' : 'default'}
                      size="small"
                    />
                    {sub.last_data && (
                      <Typography variant="caption" color="text.secondary">
                        Last data: {new Date(sub.last_data!).toLocaleTimeString()}
                      </Typography>
                    )}
                  </Box>
                </Box>
                <IconButton
                  color="error"
                  onClick={() => handleRemove(sub.symbol!)}
                  disabled={removeMutation.isPending}
                >
                  <DeleteIcon />
                </IconButton>
              </CardContent>
            </Card>
          ))}
        </Box>
      )}

      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Add Subscription</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Symbol"
            placeholder="e.g., BTC/USD, ETH/USD"
            fullWidth
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            onClick={handleAdd}
            disabled={addMutation.isPending || !symbol.trim()}
            variant="contained"
          >
            {addMutation.isPending ? <CircularProgress size={20} /> : 'Add'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
