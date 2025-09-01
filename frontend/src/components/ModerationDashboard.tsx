import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  CardHeader,
  Chip,
  LinearProgress,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Avatar,
  Button,
  IconButton,
  Tooltip,
  Alert,
  Divider,
  CircularProgress,
} from '@mui/material';
import {
  CheckCircle,
  Block,
  Warning,
  TrendingUp,
  Refresh,
  PlayArrow,
  Pause,
  Security,
  Speed,
  Timeline,
  Assessment,
} from '@mui/icons-material';
import { useConfig } from '../contexts/ConfigContext';
import { motion } from 'framer-motion';

interface ModerationStats {
  totalRequests: number;
  safeContent: number;
  blockedContent: number;
  pendingContent: number;
  averageResponseTime: number;
  successRate: number;
  topBlockedTerms: string[];
  recentActivity: Array<{
    id: string;
    type: 'post' | 'comment' | 'profile';
    content: string;
    status: 'safe' | 'blocked' | 'pending';
    timestamp: Date;
    reason?: string;
  }>;
}

const ModerationDashboard: React.FC = () => {
  const { config } = useConfig();
  const [stats, setStats] = useState<ModerationStats>({
    totalRequests: 0,
    safeContent: 0,
    blockedContent: 0,
    pendingContent: 0,
    averageResponseTime: 0,
    successRate: 0,
    topBlockedTerms: [],
    recentActivity: [],
  });
  const [isLive, setIsLive] = useState(false);
  const [loading, setLoading] = useState(false);

  // Simulate live data updates
  useEffect(() => {
    if (!isLive || !config.moderationDashboard) return;

    const interval = setInterval(() => {
      setStats(prev => ({
        ...prev,
        totalRequests: prev.totalRequests + Math.floor(Math.random() * 3),
        safeContent: prev.safeContent + Math.floor(Math.random() * 2),
        blockedContent: prev.blockedContent + Math.floor(Math.random() * 1),
        averageResponseTime: Math.max(50, prev.averageResponseTime + (Math.random() - 0.5) * 20),
        successRate: Math.min(100, Math.max(80, prev.successRate + (Math.random() - 0.5) * 5)),
      }));
    }, 3000);

    return () => clearInterval(interval);
  }, [isLive, config.moderationDashboard]);

  const generateMockData = () => {
    setLoading(true);
    setTimeout(() => {
      const mockStats: ModerationStats = {
        totalRequests: Math.floor(Math.random() * 1000) + 500,
        safeContent: Math.floor(Math.random() * 800) + 400,
        blockedContent: Math.floor(Math.random() * 200) + 50,
        pendingContent: Math.floor(Math.random() * 50) + 10,
        averageResponseTime: Math.floor(Math.random() * 200) + 100,
        successRate: Math.floor(Math.random() * 20) + 80,
        topBlockedTerms: ['spam', 'hate speech', 'inappropriate content', 'commercial', 'phishing'],
        recentActivity: [
          {
            id: '1',
            type: 'post',
            content: 'This is a great article about...',
            status: 'safe',
            timestamp: new Date(Date.now() - 5 * 60 * 1000),
          },
          {
            id: '2',
            type: 'comment',
            content: 'Buy now! Limited time offer...',
            status: 'blocked',
            timestamp: new Date(Date.now() - 10 * 60 * 1000),
            reason: 'Spam content detected',
          },
          {
            id: '3',
            type: 'profile',
            content: 'Software developer passionate about...',
            status: 'safe',
            timestamp: new Date(Date.now() - 15 * 60 * 1000),
          },
          {
            id: '4',
            type: 'post',
            content: 'I hate this content so much...',
            status: 'blocked',
            timestamp: new Date(Date.now() - 20 * 60 * 1000),
            reason: 'Hate speech detected',
          },
        ],
      };
      setStats(mockStats);
      setLoading(false);
    }, 1500);
  };

  useEffect(() => {
    if (config.moderationDashboard) {
      generateMockData();
    }
  }, [config.moderationDashboard]);

  if (!config.moderationDashboard) {
    return (
      <Alert severity="info">
        <Typography variant="body1">
          Moderation Dashboard is disabled. Enable it in the Configuration Panel to view statistics.
        </Typography>
      </Alert>
    );
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'safe': return 'success';
      case 'blocked': return 'error';
      case 'pending': return 'warning';
      default: return 'default';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'safe': return <CheckCircle color="success" />;
      case 'blocked': return <Block color="error" />;
      case 'pending': return <Warning color="warning" />;
      default: return <Warning />;
    }
  };

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
              📊 Moderation Dashboard
            </Typography>
            <Box display="flex" gap={2}>
              <Tooltip title={isLive ? 'Pause Live Updates' : 'Start Live Updates'}>
                <IconButton
                  color={isLive ? 'success' : 'default'}
                  onClick={() => setIsLive(!isLive)}
                >
                  {isLive ? <Pause /> : <PlayArrow />}
                </IconButton>
              </Tooltip>
              <Tooltip title="Refresh Data">
                <IconButton onClick={generateMockData} disabled={loading}>
                  {loading ? <CircularProgress size={24} /> : <Refresh />}
                </IconButton>
              </Tooltip>
            </Box>
          </Box>
          
          <Typography variant="body1" color="text.secondary" paragraph>
            Real-time monitoring of your content moderation system. Track performance metrics, 
            view blocked content, and monitor system health.
          </Typography>

          {isLive && (
            <Alert severity="success" sx={{ mb: 2 }}>
              <Typography variant="body2">
                🟢 Live updates enabled - Data refreshes every 3 seconds
              </Typography>
            </Alert>
          )}
        </Paper>

        {/* Key Metrics */}
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card elevation={2}>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Total Requests
                    </Typography>
                    <Typography variant="h4">
                      {stats.totalRequests.toLocaleString()}
                    </Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'primary.main' }}>
                    <Assessment />
                  </Avatar>
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card elevation={2}>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Safe Content
                    </Typography>
                    <Typography variant="h4" color="success.main">
                      {stats.safeContent.toLocaleString()}
                    </Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'success.main' }}>
                    <CheckCircle />
                  </Avatar>
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card elevation={2}>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Blocked Content
                    </Typography>
                    <Typography variant="h4" color="error.main">
                      {stats.blockedContent.toLocaleString()}
                    </Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'error.main' }}>
                    <Block />
                  </Avatar>
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} sm={6} md={3}>
            <Card elevation={2}>
              <CardContent>
                <Box display="flex" alignItems="center" justifyContent="space-between">
                  <Box>
                    <Typography color="textSecondary" gutterBottom>
                      Success Rate
                    </Typography>
                    <Typography variant="h4" color="info.main">
                      {stats.successRate.toFixed(1)}%
                    </Typography>
                  </Box>
                  <Avatar sx={{ bgcolor: 'info.main' }}>
                    <TrendingUp />
                  </Avatar>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Performance Metrics */}
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid item xs={12} md={6}>
            <Card elevation={2}>
              <CardHeader
                title="Performance Metrics"
                avatar={<Speed />}
              />
              <CardContent>
                <Box mb={3}>
                  <Box display="flex" justifyContent="space-between" mb={1}>
                    <Typography variant="body2">Response Time</Typography>
                    <Typography variant="body2">{stats.averageResponseTime.toFixed(0)}ms</Typography>
                  </Box>
                  <LinearProgress
                    variant="determinate"
                    value={Math.min(100, (stats.averageResponseTime / 300) * 100)}
                    color={stats.averageResponseTime < 150 ? 'success' : stats.averageResponseTime < 250 ? 'warning' : 'error'}
                  />
                </Box>
                
                <Box mb={3}>
                  <Box display="flex" justifyContent="space-between" mb={1}>
                    <Typography variant="body2">Content Processing</Typography>
                    <Typography variant="body2">{stats.pendingContent}</Typography>
                  </Box>
                  <LinearProgress
                    variant="determinate"
                    value={stats.pendingContent > 0 ? 50 : 0}
                    color="warning"
                  />
                </Box>

                <Box>
                  <Box display="flex" justifyContent="space-between" mb={1}>
                    <Typography variant="body2">System Health</Typography>
                    <Typography variant="body2" color="success.main">Healthy</Typography>
                  </Box>
                  <LinearProgress
                    variant="determinate"
                    value={100}
                    color="success"
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} md={6}>
            <Card elevation={2}>
              <CardHeader
                title="Top Blocked Terms"
                avatar={<Security />}
              />
              <CardContent>
                <Box display="flex" flexWrap="wrap" gap={1}>
                  {stats.topBlockedTerms.map((term, index) => (
                    <Chip
                      key={index}
                      label={term}
                      color="error"
                      variant="outlined"
                      size="small"
                    />
                  ))}
                </Box>
                <Typography variant="body2" color="textSecondary" sx={{ mt: 2 }}>
                  These terms are most commonly flagged by the moderation system
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Recent Activity */}
        <Card elevation={2}>
          <CardHeader
            title="Recent Activity"
            avatar={<Timeline />}
            action={
              <Button size="small" color="primary">
                View All
              </Button>
            }
          />
          <CardContent>
            <List>
              {stats.recentActivity.map((activity, index) => (
                <React.Fragment key={activity.id}>
                  <ListItem>
                    <ListItemIcon>
                      {getStatusIcon(activity.status)}
                    </ListItemIcon>
                    <ListItemText
                      primary={
                        <Box display="flex" alignItems="center" gap={1}>
                          <Typography variant="body1">
                            {activity.content.length > 50 
                              ? `${activity.content.substring(0, 50)}...` 
                              : activity.content
                            }
                          </Typography>
                          <Chip
                            label={activity.type}
                            size="small"
                            variant="outlined"
                          />
                          <Chip
                            label={activity.status}
                            color={getStatusColor(activity.status) as any}
                            size="small"
                          />
                        </Box>
                      }
                      secondary={
                        <Box>
                          <Typography variant="body2" color="textSecondary">
                            {activity.timestamp.toLocaleString()}
                          </Typography>
                          {activity.reason && (
                            <Typography variant="body2" color="error.main">
                              Reason: {activity.reason}
                            </Typography>
                          )}
                        </Box>
                      }
                    />
                  </ListItem>
                  {index < stats.recentActivity.length - 1 && <Divider />}
                </React.Fragment>
              ))}
            </List>
          </CardContent>
        </Card>
      </Box>
    </motion.div>
  );
};

export default ModerationDashboard;

