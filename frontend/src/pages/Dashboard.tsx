import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  CardHeader,
  CircularProgress,
  Alert,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  LinearProgress,
} from '@mui/material';
import {
  Warning,
  CheckCircle,
  Error as ErrorIcon,
  AccessTime,
  ArrowUpward,
  ArrowDownward,
  Assessment,
  Article,
  Notifications,
} from '@mui/icons-material';
import { useDashboard, useSignals, useNews, useTrades } from '../hooks/useApi';

function formatUSD(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
  }).format(value);
}

function formatTime(date: string): string {
  return new Date(date).toLocaleTimeString();
}

function PortfolioCard() {
  const { data, isLoading, error } = useDashboard();

  if (isLoading) {
    return (
      <Card>
        <CardContent sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Alert severity="error">Failed to load portfolio data</Alert>
        </CardContent>
      </Card>
    );
  }

  const portfolio = data?.portfolio;
  const balances = portfolio?.balances || {};
  const balanceEntries = Object.entries(balances).filter(([, amount]) => amount > 0);

  return (
    <Card>
      <CardHeader
        title="Portfolio"
        titleTypographyProps={{ variant: 'h6' }}
        action={
          portfolio?.updated_at && (
            <Typography variant="caption" color="text.secondary">
              Updated {formatTime(portfolio.updated_at)}
            </Typography>
          )
        }
      />
      <CardContent>
        <Box sx={{ mb: 3 }}>
          <Typography variant="h4" sx={{ fontWeight: 'bold', mb: 0.5 }}>
            {formatUSD(portfolio?.total_value_usd || 0)}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Total Value (USD)
          </Typography>
        </Box>

        {balanceEntries.length > 0 && (
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
              Balances
            </Typography>
            {balanceEntries.map(([asset, amount]) => (
              <Box
                key={asset}
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  py: 0.5,
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                }}
              >
                <Typography variant="body2">{asset}</Typography>
                <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                  {amount.toFixed(8)}
                </Typography>
              </Box>
            ))}
          </Box>
        )}

        {balanceEntries.length === 0 && (
          <Alert severity="info" sx={{ mt: 1 }}>
            No balances available. Connect your Kraken account to see portfolio data.
          </Alert>
        )}
      </CardContent>
    </Card>
  );
}

function SubscriptionHealthCard() {
  const { data, isLoading, error } = useDashboard();

  if (isLoading) {
    return (
      <Card>
        <CardContent sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Alert severity="error">Failed to load subscription health</Alert>
        </CardContent>
      </Card>
    );
  }

  const health = data?.subscription_health;
  const active = health?.active || 0;
  const stale = health?.stale || 0;
  const errored = health?.errored || 0;
  const errors = health?.errors || {};
  const total = active + stale;

  return (
    <Card>
      <CardHeader title="Subscription Health" titleTypographyProps={{ variant: 'h6' }} />
      <CardContent>
        <Grid container spacing={2} sx={{ mb: 2 }}>
          <Grid size={4}>
            <Box sx={{ textAlign: 'center' }}>
              <CheckCircle color="success" sx={{ fontSize: 32 }} />
              <Typography variant="h5">{active}</Typography>
              <Typography variant="caption" color="text.secondary">Active</Typography>
            </Box>
          </Grid>
          <Grid size={4}>
            <Box sx={{ textAlign: 'center' }}>
              <AccessTime color="warning" sx={{ fontSize: 32 }} />
              <Typography variant="h5">{stale}</Typography>
              <Typography variant="caption" color="text.secondary">Stale</Typography>
            </Box>
          </Grid>
          <Grid size={4}>
            <Box sx={{ textAlign: 'center' }}>
              <ErrorIcon color="error" sx={{ fontSize: 32 }} />
              <Typography variant="h5">{errored}</Typography>
              <Typography variant="caption" color="text.secondary">Errors</Typography>
            </Box>
          </Grid>
        </Grid>

        {total > 0 && (
          <Box sx={{ mb: 2 }}>
            <Box sx={{ display: 'flex', mb: 0.5 }}>
              <Typography variant="caption" color="text.secondary">
                Data freshness
              </Typography>
            </Box>
            <Box sx={{ display: 'flex', height: 8, borderRadius: 4, overflow: 'hidden', bgcolor: 'grey.200' }}>
              {active > 0 && (
                <Box
                  sx={{
                    width: `${(active / total) * 100}%`,
                    bgcolor: 'success.main',
                  }}
                />
              )}
              {stale > 0 && (
                <Box
                  sx={{
                    width: `${(stale / total) * 100}%`,
                    bgcolor: 'warning.main',
                  }}
                />
              )}
            </Box>
          </Box>
        )}

        {Object.keys(errors).length > 0 && (
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
              Recent Errors
            </Typography>
            {Object.entries(errors).slice(0, 3).map(([symbol, message]) => (
              <Alert key={symbol} severity="error" sx={{ mb: 1 }} icon={<Warning />}>
                <Typography variant="caption">
                  <strong>{symbol}</strong>: {message}
                </Typography>
              </Alert>
            ))}
          </Box>
        )}
      </CardContent>
    </Card>
  );
}

function MarketOverviewCard() {
  const { data, isLoading, error } = useDashboard();

  if (isLoading) {
    return (
      <Card>
        <CardContent sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Alert severity="error">Failed to load market data</Alert>
        </CardContent>
      </Card>
    );
  }

  const snapshots = data?.market_snapshot || [];

  return (
    <Card>
      <CardHeader title="Market Overview" titleTypographyProps={{ variant: 'h6' }} />
      <CardContent>
        {snapshots.length === 0 ? (
          <Alert severity="info">
            No market data available. Add subscriptions to see live prices.
          </Alert>
        ) : (
          <TableContainer component={Paper} variant="outlined">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Pair</TableCell>
                  <TableCell align="right">Last</TableCell>
                  <TableCell align="right">Bid</TableCell>
                  <TableCell align="right">Ask</TableCell>
                  <TableCell align="right">Volume 24h</TableCell>
                  <TableCell align="center">Updated</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {snapshots.map((snap) => (
                  <TableRow key={snap.pair}>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 'medium' }}>
                        {snap.pair}
                      </Typography>
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {snap.last?.toFixed(2)}
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {snap.bid?.toFixed(2)}
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {snap.ask?.toFixed(2)}
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {snap.volume_24h?.toFixed(4)}
                    </TableCell>
                    <TableCell align="center">
                      {snap.updated_at && (
                        <Chip
                          size="small"
                          label={formatTime(snap.updated_at)}
                          color={
                            Date.now() - new Date(snap.updated_at).getTime() < 30000
                              ? 'success'
                              : 'default'
                          }
                          variant="outlined"
                        />
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </CardContent>
    </Card>
  );
}

function RecentDecisionsCard() {
  const { data, isLoading, error } = useDashboard();

  if (isLoading) {
    return (
      <Card>
        <CardContent sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Alert severity="error">Failed to load decisions</Alert>
        </CardContent>
      </Card>
    );
  }

  const decisions = data?.recent_decisions || [];

  return (
    <Card>
      <CardHeader
        title="Recent LLM Decisions"
        titleTypographyProps={{ variant: 'h6' }}
        action={
          <Chip
            size="small"
            label={`${decisions.length} recent`}
            variant="outlined"
          />
        }
      />
      <CardContent>
        {decisions.length === 0 ? (
          <Alert severity="info">
            No decisions yet. Price alerts will trigger LLM analysis.
          </Alert>
        ) : (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Time</TableCell>
                  <TableCell>Pair</TableCell>
                  <TableCell>Action</TableCell>
                  <TableCell>Confidence</TableCell>
                  <TableCell>Size %</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {decisions.map((decision) => (
                  <TableRow key={decision.id}>
                    <TableCell>
                      {decision.created_at && formatTime(decision.created_at)}
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 'medium' }}>
                        {decision.pair}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={decision.action?.toUpperCase()}
                        color={
                          decision.action === 'buy'
                            ? 'success'
                            : decision.action === 'sell'
                            ? 'error'
                            : 'default'
                        }
                      />
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <LinearProgress
                          variant="determinate"
                          value={(decision.confidence || 0) * 10}
                          sx={{ width: 60, height: 6, borderRadius: 3 }}
                          color={
                            (decision.confidence || 0) > 0.7
                              ? 'success'
                              : (decision.confidence || 0) > 0.4
                              ? 'warning'
                              : 'error'
                          }
                        />
                        <Typography variant="caption">
                          {((decision.confidence || 0) * 100).toFixed(0)}%
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell sx={{ fontFamily: 'monospace' }}>
                      {decision.size_pct?.toFixed(1)}%
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </CardContent>
    </Card>
  );
}

function SignalsCard() {
  const { data, isLoading, error } = useSignals();

  if (isLoading) {
    return (
      <Card>
        <CardContent sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Alert severity="error">Failed to load signals</Alert>
        </CardContent>
      </Card>
    );
  }

  const signals = data?.signals || [];

  return (
    <Card>
      <CardHeader title="PRISM Signals" titleTypographyProps={{ variant: 'h6' }} />
      <CardContent>
        {signals.length === 0 ? (
          <Alert severity="info">
            No technical signals available. PRISM analysis runs periodically.
          </Alert>
        ) : (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Symbol</TableCell>
                  <TableCell>Momentum</TableCell>
                  <TableCell>Breakout</TableCell>
                  <TableCell>Volume</TableCell>
                  <TableCell>Updated</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {signals.map((signal) => (
                  <TableRow key={signal.symbol}>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 'medium' }}>
                        {signal.symbol}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={signal.momentum_signal || '—'}
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={signal.breakout_signal || '—'}
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={signal.volume_signal || '—'}
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell>
                      {signal.updated_at && (
                        <Typography variant="caption" color="text.secondary">
                          {formatTime(signal.updated_at)}
                        </Typography>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </CardContent>
    </Card>
  );
}

function PriceAlertsCard() {
  const { data, isLoading, error } = useDashboard();

  if (isLoading) {
    return (
      <Card>
        <CardContent sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Alert severity="error">Failed to load alerts</Alert>
        </CardContent>
      </Card>
    );
  }

  const alerts = data?.recent_alerts || [];

  return (
    <Card>
      <CardHeader
        title="Price Alerts"
        titleTypographyProps={{ variant: 'h6' }}
        action={<Notifications fontSize="small" color="action" />}
      />
      <CardContent>
        {alerts.length === 0 ? (
          <Alert severity="info">
            No recent price alerts. Alerts trigger on &gt;0.5% price movements.
          </Alert>
        ) : (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Pair</TableCell>
                  <TableCell>Change</TableCell>
                  <TableCell align="right">From</TableCell>
                  <TableCell align="right">To</TableCell>
                  <TableCell>Time</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {alerts.map((alert, idx) => (
                  <TableRow key={idx}>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 'medium' }}>
                        {alert.pair}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        icon={
                          (alert.change_pct || 0) >= 0 ? (
                            <ArrowUpward fontSize="small" />
                          ) : (
                            <ArrowDownward fontSize="small" />
                          )
                        }
                        label={`${(alert.change_pct || 0) >= 0 ? '+' : ''}${(alert.change_pct || 0).toFixed(2)}%`}
                        color={(alert.change_pct || 0) >= 0 ? 'success' : 'error'}
                      />
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {alert.previous?.toFixed(2)}
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {alert.current?.toFixed(2)}
                    </TableCell>
                    <TableCell>
                      {alert.updated_at && (
                        <Typography variant="caption" color="text.secondary">
                          {formatTime(alert.updated_at)}
                        </Typography>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </CardContent>
    </Card>
  );
}

function RecentTradesCard() {
  const { data, isLoading, error } = useTrades();

  if (isLoading) {
    return (
      <Card>
        <CardContent sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Alert severity="error">Failed to load trades</Alert>
        </CardContent>
      </Card>
    );
  }

  const trades = data?.trades || [];

  return (
    <Card>
      <CardHeader
        title="Trade History"
        titleTypographyProps={{ variant: 'h6' }}
        action={<Assessment fontSize="small" color="action" />}
      />
      <CardContent>
        {trades.length === 0 ? (
          <Alert severity="info">
            No trade history. Trades will appear when the bot executes orders.
          </Alert>
        ) : (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Time</TableCell>
                  <TableCell>Pair</TableCell>
                  <TableCell>Action</TableCell>
                  <TableCell>Mode</TableCell>
                  <TableCell align="right">Price</TableCell>
                  <TableCell align="right">Size</TableCell>
                  <TableCell align="right">Cost</TableCell>
                  <TableCell>Confidence</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {trades.map((trade) => (
                  <TableRow key={trade.timestamp}>
                    <TableCell>
                      {trade.timestamp && formatTime(trade.timestamp)}
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 'medium' }}>
                        {trade.pair}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={trade.action?.toUpperCase()}
                        color={
                          trade.action === 'buy'
                            ? 'success'
                            : trade.action === 'sell'
                            ? 'error'
                            : 'default'
                        }
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={trade.mode || '—'}
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {trade.price?.toFixed(2)}
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {trade.size?.toFixed(6)}
                    </TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace' }}>
                      {trade.cost?.toFixed(2)}
                    </TableCell>
                    <TableCell>
                      {trade.confidence != null && (
                        <Typography variant="caption">
                          {(trade.confidence * 100).toFixed(0)}%
                        </Typography>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </CardContent>
    </Card>
  );
}

export function DashboardPage() {
  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 3 }}>
        Dashboard
      </Typography>

      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 4 }}>
          <PortfolioCard />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <SubscriptionHealthCard />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <MarketOverviewCard />
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <RecentDecisionsCard />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <SignalsCard />
        </Grid>

        <Grid size={{ xs: 12 }}>
          <RecentTradesCard />
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <PriceAlertsCard />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <NewsCard />
        </Grid>
      </Grid>
    </Box>
  );
}

function NewsCard() {
  const { data, isLoading, error } = useNews();

  if (isLoading) {
    return (
      <Card>
        <CardContent sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Alert severity="error">Failed to load news</Alert>
        </CardContent>
      </Card>
    );
  }

  const news = data?.news || [];

  return (
    <Card>
      <CardHeader
        title="Recent News"
        titleTypographyProps={{ variant: 'h6' }}
        action={<Article fontSize="small" color="action" />}
      />
      <CardContent>
        {news.length === 0 ? (
          <Alert severity="info">
            No news articles available. News crawler runs periodically.
          </Alert>
        ) : (
          <Box>
            {news.slice(0, 5).map((item) => (
              <Box
                key={item.id}
                sx={{
                  pb: 2,
                  mb: 2,
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                }}
              >
                <Typography variant="body2" sx={{ fontWeight: 'medium', mb: 0.5 }}>
                  {item.title}
                </Typography>
                <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
                  <Chip size="small" label={item.source} variant="outlined" />
                  {item.timestamp && (
                    <Typography variant="caption" color="text.secondary">
                      {formatTime(item.timestamp)}
                    </Typography>
                  )}
                </Box>
              </Box>
            ))}
          </Box>
        )}
      </CardContent>
    </Card>
  );
}
