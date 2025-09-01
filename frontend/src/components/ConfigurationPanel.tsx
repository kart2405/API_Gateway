import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  Switch,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
  Button,
  Alert,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  InputAdornment,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  ExpandMore,
  Palette,
  Tune,
  Security,
  Code,
  Cloud,
  Save,
  Refresh,
  Preview,
  CheckCircle,
  Warning,
  Info,
} from '@mui/icons-material';
import { useConfig } from '../contexts/ConfigContext';
import { motion } from 'framer-motion';

const ConfigurationPanel: React.FC = () => {
  const { config, updateConfig, resetConfig, saveConfig } = useConfig();
  const [activeSection, setActiveSection] = useState<string | false>('ui');
  const [showPreview, setShowPreview] = useState(false);

  const handleChange = (key: keyof typeof config, value: any) => {
    updateConfig({ [key]: value });
  };

  const handleNestedChange = (parentKey: string, childKey: string, value: any) => {
    const parentConfig = config[parentKey as keyof typeof config] as any;
    if (parentConfig && typeof parentConfig === 'object') {
      updateConfig({
        [parentKey]: {
          ...parentConfig,
          [childKey]: value,
        },
      });
    }
  };

  const sections = [
    {
      id: 'ui',
      title: '🎨 UI Style & Theme',
      icon: <Palette />,
      content: (
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <FormControl fullWidth>
              <InputLabel>Theme Style</InputLabel>
              <Select
                value={config.theme}
                label="Theme Style"
                onChange={(e) => handleChange('theme', e.target.value)}
              >
                <MenuItem value="modern">Modern Minimal</MenuItem>
                <MenuItem value="colorful">Colorful Creative</MenuItem>
                <MenuItem value="dark">Dark Mode</MenuItem>
                <MenuItem value="light">Light Mode</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <FormControl fullWidth>
              <InputLabel>Color Scheme</InputLabel>
              <Select
                value={config.colorScheme}
                label="Color Scheme"
                onChange={(e) => handleChange('colorScheme', e.target.value)}
              >
                <MenuItem value="blue">Blue</MenuItem>
                <MenuItem value="green">Green</MenuItem>
                <MenuItem value="purple">Purple</MenuItem>
                <MenuItem value="custom">Custom</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <FormControl fullWidth>
              <InputLabel>Layout Style</InputLabel>
              <Select
                value={config.layout}
                label="Layout Style"
                onChange={(e) => handleChange('layout', e.target.value)}
              >
                <MenuItem value="compact">Compact</MenuItem>
                <MenuItem value="spacious">Spacious</MenuItem>
                <MenuItem value="sidebar">Sidebar</MenuItem>
                <MenuItem value="top">Top Navigation</MenuItem>
              </Select>
            </FormControl>
          </Grid>

          {config.colorScheme === 'custom' && (
            <Grid item xs={12}>
              <Typography variant="h6" gutterBottom>
                Custom Colors
              </Typography>
              <Grid container spacing={2}>
                <Grid item xs={12} md={4}>
                  <TextField
                    fullWidth
                    label="Primary Color"
                    type="color"
                    value={config.customColors.primary}
                    onChange={(e) => handleNestedChange('customColors', 'primary', e.target.value)}
                    InputProps={{
                      startAdornment: <InputAdornment position="start">🎨</InputAdornment>,
                    }}
                  />
                </Grid>
                <Grid item xs={12} md={4}>
                  <TextField
                    fullWidth
                    label="Secondary Color"
                    type="color"
                    value={config.customColors.secondary}
                    onChange={(e) => handleNestedChange('customColors', 'secondary', e.target.value)}
                    InputProps={{
                      startAdornment: <InputAdornment position="start">🎨</InputAdornment>,
                    }}
                  />
                </Grid>
                <Grid item xs={12} md={4}>
                  <TextField
                    fullWidth
                    label="Accent Color"
                    type="color"
                    value={config.customColors.accent}
                    onChange={(e) => handleNestedChange('customColors', 'accent', e.target.value)}
                    InputProps={{
                      startAdornment: <InputAdornment position="start">🎨</InputAdornment>,
                    }}
                  />
                </Grid>
              </Grid>
            </Grid>
          )}
        </Grid>
      ),
    },
    {
      id: 'features',
      title: '🚀 Features & Functionality',
      icon: <Tune />,
      content: (
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Real-time Testing</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Live API endpoint testing with instant results
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.realTimeTesting}
                    onChange={(e) => handleChange('realTimeTesting', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Moderation Dashboard</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Visual dashboard showing moderation statistics
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.moderationDashboard}
                    onChange={(e) => handleChange('moderationDashboard', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Content Forms</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Interactive forms for testing different content types
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.contentForms}
                    onChange={(e) => handleChange('contentForms', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Statistics & Analytics</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Detailed metrics and performance data
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.statistics}
                    onChange={(e) => handleChange('statistics', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">User Management</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Advanced user authentication and management
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.userManagement}
                    onChange={(e) => handleChange('userManagement', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      ),
    },
    {
      id: 'auth',
      title: '🔐 Authentication & Security',
      icon: <Security />,
      content: (
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Demo Mode</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Use pre-generated demo tokens for testing
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.demoMode}
                    onChange={(e) => handleChange('demoMode', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Login Form</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Interactive login form for real authentication
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.loginForm}
                    onChange={(e) => handleChange('loginForm', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">User Registration</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Allow new users to create accounts
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.userRegistration}
                    onChange={(e) => handleChange('userRegistration', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      ),
    },
    {
      id: 'content',
      title: '📝 Content Testing Options',
      icon: <Code />,
      content: (
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Post Testing</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Test blog post content moderation
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.postTesting}
                    onChange={(e) => handleChange('postTesting', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Comment Testing</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Test comment content moderation
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.commentTesting}
                    onChange={(e) => handleChange('commentTesting', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Profile Testing</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Test user profile content moderation
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.profileTesting}
                    onChange={(e) => handleChange('profileTesting', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Custom Content</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Test arbitrary content with custom validation
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.customContent}
                    onChange={(e) => handleChange('customContent', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      ),
    },
    {
      id: 'deployment',
      title: '🌐 Deployment & API Settings',
      icon: <Cloud />,
      content: (
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Frontend Port"
              type="number"
              value={config.port}
              onChange={(e) => handleChange('port', parseInt(e.target.value))}
              InputProps={{
                startAdornment: <InputAdornment position="start">🔌</InputAdornment>,
              }}
            />
          </Grid>
          
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="API Endpoint"
              value={config.apiEndpoint}
              onChange={(e) => handleChange('apiEndpoint', e.target.value)}
              InputProps={{
                startAdornment: <InputAdornment position="start">🌐</InputAdornment>,
              }}
            />
          </Grid>
          
          <Grid item xs={12} md={6}>
            <Card variant="outlined">
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography variant="h6">Proxy Settings</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Handle CORS and proxy configuration
                    </Typography>
                  </Box>
                  <Switch
                    checked={config.proxySettings}
                    onChange={(e) => handleChange('proxySettings', e.target.checked)}
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      ),
    },
  ];

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <Box>
        <Paper elevation={3} sx={{ p: 3, mb: 3 }}>
          <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
            <Typography variant="h4" component="h1">
              🎛️ Configuration Panel
            </Typography>
            <Box>
              <Tooltip title="Preview Configuration">
                <IconButton
                  color="primary"
                  onClick={() => setShowPreview(!showPreview)}
                >
                  <Preview />
                </IconButton>
              </Tooltip>
              <Tooltip title="Save Configuration">
                <IconButton color="success" onClick={saveConfig}>
                  <Save />
                </IconButton>
              </Tooltip>
              <Tooltip title="Reset to Defaults">
                <IconButton color="warning" onClick={resetConfig}>
                  <Refresh />
                </IconButton>
              </Tooltip>
            </Box>
          </Box>
          
          <Typography variant="body1" color="text.secondary" paragraph>
            Customize your moderation system experience! Choose your preferred UI style, 
            enable/disable features, configure authentication, and set up deployment options.
          </Typography>

          {showPreview && (
            <Alert severity="info" sx={{ mb: 2 }}>
              <Typography variant="body2">
                <strong>Configuration Preview:</strong> {config.theme} theme, {config.colorScheme} colors, 
                {config.realTimeTesting ? ' real-time testing enabled' : ' real-time testing disabled'}, 
                {config.moderationDashboard ? ' dashboard enabled' : ' dashboard disabled'}
              </Typography>
            </Alert>
          )}
        </Paper>

        <Grid container spacing={3}>
          {sections.map((section) => (
            <Grid item xs={12} key={section.id}>
              <Accordion
                expanded={activeSection === section.id}
                onChange={() => setActiveSection(activeSection === section.id ? false : section.id)}
                sx={{ 
                  '&:before': { display: 'none' },
                  boxShadow: 3,
                  borderRadius: 2,
                }}
              >
                <AccordionSummary
                  expandIcon={<ExpandMore />}
                  sx={{
                    backgroundColor: 'primary.main',
                    color: 'white',
                    borderRadius: activeSection === section.id ? '8px 8px 0 0' : '8px',
                    '&:hover': {
                      backgroundColor: 'primary.dark',
                    },
                  }}
                >
                  <Box display="flex" alignItems="center" gap={2}>
                    {section.icon}
                    <Typography variant="h6">{section.title}</Typography>
                  </Box>
                </AccordionSummary>
                <AccordionDetails sx={{ p: 3 }}>
                  {section.content}
                </AccordionDetails>
              </Accordion>
            </Grid>
          ))}
        </Grid>

        <Paper elevation={3} sx={{ p: 3, mt: 3 }}>
          <Typography variant="h6" gutterBottom>
            🎯 Quick Actions
          </Typography>
          <Box display="flex" gap={2} flexWrap="wrap">
            <Button
              variant="contained"
              color="primary"
              startIcon={<CheckCircle />}
              onClick={() => {
                updateConfig({
                  theme: 'modern',
                  colorScheme: 'blue',
                  realTimeTesting: true,
                  moderationDashboard: true,
                });
              }}
            >
              Apply Modern Blue Theme
            </Button>
            <Button
              variant="contained"
              color="secondary"
              startIcon={<CheckCircle />}
              onClick={() => {
                updateConfig({
                  theme: 'colorful',
                  colorScheme: 'purple',
                  realTimeTesting: true,
                  contentForms: true,
                });
              }}
            >
              Apply Colorful Purple Theme
            </Button>
            <Button
              variant="outlined"
              color="info"
              startIcon={<Info />}
              onClick={() => {
                updateConfig({
                  demoMode: true,
                  loginForm: false,
                  userRegistration: false,
                });
              }}
            >
              Demo Mode Only
            </Button>
            <Button
              variant="outlined"
              color="warning"
              startIcon={<Warning />}
              onClick={() => {
                updateConfig({
                  realTimeTesting: false,
                  statistics: false,
                  userManagement: false,
                });
              }}
            >
              Minimal Features
            </Button>
          </Box>
        </Paper>
      </Box>
    </motion.div>
  );
};

export default ConfigurationPanel;
