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
  CircularProgress,
  Alert,
  Autocomplete,
  TextField,
} from '@mui/material';
import { Add as AddIcon, Delete as DeleteIcon } from '@mui/icons-material';
import {
  useSubscriptionDetails,
  useAddSubscription,
  useRemoveSubscription,
  usePairs,
} from '../hooks/useApi';

export function SubscriptionsPage() {
  const { data, isLoading, error } = useSubscriptionDetails();
  const { data: pairsData } = usePairs();
  const addMutation = useAddSubscription();
  const removeMutation = useRemoveSubscription();
  const [open, setOpen] = useState(false);
  const [selectedPair, setSelectedPair] = useState<string | null>(null);

  const handleAdd = () => {
    if (selectedPair) {
      addMutation.mutate(selectedPair, {
        onSuccess: () => {
          setOpen(false);
          setSelectedPair(null);
        },
      });
    }
  };

  const handleRemove = (sym: string) => {
    if (confirm(`Remove subscription for ${sym}?`)) {
      removeMutation.mutate(sym);
    }
  };

  const pairs = pairsData?.pairs?.map(p => p.symbol || p.altname || p.ws_name || '') || [];

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

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Subscription</DialogTitle>
        <DialogContent>
          <Autocomplete
            options={pairs}
            value={selectedPair}
            onChange={(_, value) => setSelectedPair(value)}
            renderInput={(params) => (
              <TextField
                {...params}
                autoFocus
                margin="dense"
                label="Trading Pair"
                placeholder="Select a pair..."
                fullWidth
              />
            )}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            onClick={handleAdd}
            disabled={addMutation.isPending || !selectedPair}
            variant="contained"
          >
            {addMutation.isPending ? <CircularProgress size={20} /> : 'Add'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
