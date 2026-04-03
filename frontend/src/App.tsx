import { HashRouter, Routes, Route, NavLink } from 'react-router-dom';
import {
  Box,
  AppBar,
  Toolbar,
  Typography,
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Container,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Subscriptions as SubscriptionsIcon,
  AccountBalance as AssetsIcon,
  History as PromptsIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';

import { SubscriptionsPage } from './pages/Subscriptions';
import { AssetsPage } from './pages/Assets';
import { PromptsPage } from './pages/Prompts';
import { SettingsPage } from './pages/Settings';

const drawerWidth = 220;

function DashboardPlaceholder() {
  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 3 }}>Dashboard</Typography>
      <Typography color="text.secondary">
        Welcome to Kraken Trader. Use the navigation to manage subscriptions, view assets, and monitor prompts.
      </Typography>
    </Box>
  );
}

const navItems = [
  { path: '/', label: 'Dashboard', icon: <DashboardIcon /> },
  { path: '/subscriptions', label: 'Subscriptions', icon: <SubscriptionsIcon /> },
  { path: '/assets', label: 'Assets', icon: <AssetsIcon /> },
  { path: '/prompts', label: 'Prompts', icon: <PromptsIcon /> },
  { path: '/settings', label: 'Settings', icon: <SettingsIcon /> },
];

export default function App() {
  return (
    <HashRouter>
      <Box sx={{ display: 'flex' }}>
        <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
          <Toolbar>
            <Typography variant="h6" noWrap component="div">
              Kraken Trader
            </Typography>
          </Toolbar>
        </AppBar>
        <Drawer
          variant="permanent"
          sx={{
            width: drawerWidth,
            flexShrink: 0,
            '& .MuiDrawer-paper': {
              width: drawerWidth,
              boxSizing: 'border-box',
            },
          }}
        >
          <Toolbar />
          <Box sx={{ overflow: 'auto' }}>
            <List>
              {navItems.map((item) => (
                <ListItem key={item.path} disablePadding>
                  <ListItemButton component={NavLink} to={item.path}>
                    <ListItemIcon>{item.icon}</ListItemIcon>
                    <ListItemText primary={item.label} />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
          </Box>
        </Drawer>
        <Box
          component="main"
          sx={{ flexGrow: 1, p: 3, width: { sm: `calc(100% - ${drawerWidth}px)` } }}
        >
          <Toolbar />
          <Container maxWidth="lg">
            <Routes>
              <Route path="/" element={<DashboardPlaceholder />} />
              <Route path="/subscriptions" element={<SubscriptionsPage />} />
              <Route path="/assets" element={<AssetsPage />} />
              <Route path="/prompts" element={<PromptsPage />} />
              <Route path="/settings" element={<SettingsPage />} />
            </Routes>
          </Container>
        </Box>
      </Box>
    </HashRouter>
  );
}
