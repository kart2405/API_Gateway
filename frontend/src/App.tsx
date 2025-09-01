import React, { useState } from 'react';
import { 
  ThemeProvider, 
  createTheme, 
  CssBaseline, 
  Box, 
  Container,
  Typography,
  AppBar,
  Toolbar,
  IconButton,
  useMediaQuery
} from '@mui/material';
import { 
  Brightness4, 
  Brightness7, 
  Settings,
  Dashboard,
  Security,
  Code
} from '@mui/icons-material';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import ConfigurationPanel from './components/ConfigurationPanel';
import ModerationDashboard from './components/ModerationDashboard';
import APITesting from './components/APITesting';
import { ConfigProvider } from './contexts/ConfigContext';

function App() {
  const [darkMode, setDarkMode] = useState(false);
  const [currentView, setCurrentView] = useState('config');
  const isMobile = useMediaQuery('(max-width:600px)');

  const theme = createTheme({
    palette: {
      mode: darkMode ? 'dark' : 'light',
      primary: {
        main: '#1976d2',
      },
      secondary: {
        main: '#dc004e',
      },
    },
    components: {
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: 'none',
          },
        },
      },
    },
  });

  const toggleDarkMode = () => {
    setDarkMode(!darkMode);
  };

  const navigationItems = [
    { id: 'config', label: 'Configuration', icon: <Settings />, path: '/' },
    { id: 'dashboard', label: 'Dashboard', icon: <Dashboard />, path: '/dashboard' },
    { id: 'testing', label: 'API Testing', icon: <Code />, path: '/testing' },
  ];

  return (
    <Router>
      <ConfigProvider>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          <Box sx={{ flexGrow: 1 }}>
            <AppBar position="static" elevation={0}>
              <Toolbar>
                <Security sx={{ mr: 2 }} />
                <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
                  API Gateway Moderation System
                </Typography>
                
                {!isMobile && (
                  <Box sx={{ display: 'flex', gap: 1, mr: 2 }}>
                    {navigationItems.map((item) => (
                      <Link
                        key={item.id}
                        to={item.path}
                        style={{ textDecoration: 'none', color: 'inherit' }}
                      >
                        <IconButton
                          color="inherit"
                          onClick={() => setCurrentView(item.id)}
                          sx={{
                            backgroundColor: currentView === item.id ? 'rgba(255,255,255,0.1)' : 'transparent',
                            '&:hover': {
                              backgroundColor: 'rgba(255,255,255,0.2)',
                            },
                          }}
                        >
                          {item.icon}
                        </IconButton>
                      </Link>
                    ))}
                  </Box>
                )}
                
                <IconButton color="inherit" onClick={toggleDarkMode}>
                  {darkMode ? <Brightness7 /> : <Brightness4 />}
                </IconButton>
              </Toolbar>
            </AppBar>

            <Container maxWidth="xl" sx={{ mt: 3, mb: 3 }}>
              <Routes>
                <Route path="/" element={<ConfigurationPanel />} />
                <Route path="/dashboard" element={<ModerationDashboard />} />
                <Route path="/testing" element={<APITesting />} />
              </Routes>
            </Container>
          </Box>
        </ThemeProvider>
      </ConfigProvider>
    </Router>
  );
}

export default App;
