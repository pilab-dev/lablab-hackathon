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
import { Add as AddIcon, Delete as DeleteIcon, Warning as WarningIcon } from '@mui/icons-material';
import {
  useSubscriptions,
  useAddSubscription,
  useRemoveSubscription,
  usePairs,
  useTicker,
} from '../hooks/useApi';
import { PriceChart } from '../components/PriceChart';

export function SubscriptionsPage() {
  const { data, isLoading, error } = useSubscriptions();
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

  const pairs = pairsData?.pairs?.map(p => p.ws_name || p.symbol || p.altname || '') || [];

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
        <Alert severity="info">No active subscriptions. Add one to start receiving market data.</Alert>
      ) : (
        <Box sx={{ display: 'grid', gap: 2 }}>
          {data?.subscriptions?.map((sub) => (
            <SubscriptionCard
              key={sub.symbol}
              symbol={sub.symbol!}
              isActive={sub.is_active || false}
              lastData={sub.last_data}
              lastError={sub.last_error}
              onRemove={() => handleRemove(sub.symbol!)}
              isRemoving={removeMutation.isPending}
            />
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

function SubscriptionCard({
  symbol,
  isActive,
  lastData,
  lastError,
  onRemove,
  isRemoving,
}: {
  symbol: string;
  isActive: boolean;
  lastData?: string | null;
  lastError?: string | null;
  onRemove: () => void;
  isRemoving: boolean;
}) {
  const { data: tickerData } = useTicker(isActive ? symbol : '');

  const isStale = lastData ? Date.now() - new Date(lastData).getTime() > 30000 : false;

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Box>
            <Typography variant="h6">{symbol}</Typography>
            <Box sx={{ display: 'flex', gap: 1, mt: 1, alignItems: 'center' }}>
              <Chip
                label={isActive ? 'Active' : 'Inactive'}
                color={isActive ? 'success' : 'default'}
                size="small"
              />
              {isStale && (
                <Chip
                  label="Stale"
                  color="warning"
                  size="small"
                  icon={<WarningIcon />}
                />
              )}
              {lastData && (
                <Typography variant="caption" color="text.secondary">
                  Last data: {new Date(lastData).toLocaleTimeString()}
                </Typography>
              )}
            </Box>
            {lastError && (
              <Alert severity="error" sx={{ mt: 1 }} icon={<WarningIcon />}>
                <Typography variant="caption">{lastError}</Typography>
              </Alert>
            )}
            {isActive && tickerData && (
              <Box sx={{ display: 'flex', gap: 2, mt: 1 }}>
                <Typography variant="body2" color="text.secondary">
                  Bid: <Typography component="span" sx={{ fontFamily: 'monospace' }}>
                    {tickerData.bid?.toFixed(2)}
                  </Typography>
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Ask: <Typography component="span" sx={{ fontFamily: 'monospace' }}>
                    {tickerData.ask?.toFixed(2)}
                  </Typography>
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Last: <Typography component="span" sx={{ fontFamily: 'monospace' }}>
                    {tickerData.last?.toFixed(2)}
                  </Typography>
                </Typography>
              </Box>
            )}
          </Box>
          <IconButton
            color="error"
            onClick={onRemove}
            disabled={isRemoving}
          >
            <DeleteIcon />
          </IconButton>
        </Box>
        {isActive && symbol && (
          <PriceChart symbol={symbol} height={200} />
        )}
      </CardContent>
    </Card>
  );
}
