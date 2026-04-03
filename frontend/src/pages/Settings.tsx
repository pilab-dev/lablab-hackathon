import {
  Box,
  Card,
  CardContent,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
} from '@mui/material';
import { useLogLevel, useSetLogLevel } from '../hooks/useApi';

const LOG_LEVELS = ['trace', 'debug', 'info', 'warn', 'error'];

export function SettingsPage() {
  const { data, isLoading } = useLogLevel();
  const setLogLevel = useSetLogLevel();

  if (isLoading) return <CircularProgress />;

  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 3 }}>Settings</Typography>
      
      <Card sx={{ maxWidth: 400 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Log Level</Typography>
          <FormControl fullWidth>
            <InputLabel>Level</InputLabel>
            <Select
              value={data?.level || 'info'}
              label="Level"
              onChange={(e) => setLogLevel.mutate(e.target.value)}
              disabled={setLogLevel.isPending}
            >
              {LOG_LEVELS.map((level) => (
                <MenuItem key={level} value={level}>
                  {level.toUpperCase()}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </CardContent>
      </Card>
    </Box>
  );
}
