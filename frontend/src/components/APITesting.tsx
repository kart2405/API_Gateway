import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  CardHeader,
  TextField,
  Button,
  Chip,
  Alert,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Divider,
  CircularProgress,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  IconButton,
  Tooltip,
  Snackbar,
} from '@mui/material';
import {
  Send,
  CheckCircle,
  Block,
  Warning,
  Refresh,
  PlayArrow,
  Stop,
  Code,
  History,
  ContentCopy,
} from '@mui/icons-material';
import { useConfig } from '../contexts/ConfigContext';
import { motion } from 'framer-motion';

interface TestResult {
  id: string;
  endpoint: string;
  method: string;
  requestBody: any;
  response: any;
  status: number;
  timestamp: Date;
  duration: number;
  moderationResult?: {
    isSafe: boolean;
    reason?: string;
  };
}

interface TestForm {
  title?: string;
  content: string;
  tags?: string[];
  postId?: number;
  bio?: string;
  website?: string;
  location?: string;
}

const APITesting: React.FC = () => {
  const { config } = useConfig();
  const [testResults, setTestResults] = useState<TestResult[]>([]);
  const [isRunning, setIsRunning] = useState(false);
  const [currentTest, setCurrentTest] = useState<string | null>(null);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' | 'warning' }>({
    open: false,
    message: '',
    severity: 'info',
  });

  // Form states
  const [postForm, setPostForm] = useState<TestForm>({
    title: '',
    content: '',
    tags: [],
  });
  const [commentForm, setCommentForm] = useState<TestForm>({
    content: '',
    postId: 123,
  });
  const [profileForm, setProfileForm] = useState<TestForm>({
    content: '',
    bio: '',
    website: '',
    location: '',
  });

  const [customForm, setCustomForm] = useState({
    endpoint: '/api/custom',
    method: 'POST',
    content: '',
  });

  // Demo content examples
  const demoContent = {
    safe: {
      title: 'Introduction to Go Programming',
      content: 'Go is a wonderful programming language that makes concurrent programming simple and efficient.',
      tags: ['golang', 'tutorial', 'programming'],
    },
    unsafe: {
      title: 'Spam Post',
      content: 'Buy now! Limited time offer. Click here to get rich quick with our amazing product.',
      tags: ['spam', 'commercial'],
    },
    hate: {
      title: 'Negative Post',
      content: 'I hate this content so much. This is terrible and should be removed immediately.',
      tags: ['negative', 'hate'],
    },
  };

  const handleTest = async (endpoint: string, method: string, body: any, testType: string) => {
    if (!config.realTimeTesting) {
      setSnackbar({
        open: true,
        message: 'Real-time testing is disabled. Enable it in the Configuration Panel.',
        severity: 'warning',
      });
      return;
    }

    const testId = `${testType}_${Date.now()}`;
    setCurrentTest(testId);
    const startTime = Date.now();

    try {
      // Simulate API call with moderation
      const response = await simulateAPICall(endpoint, method, body);
      const duration = Date.now() - startTime;

      const result: TestResult = {
        id: testId,
        endpoint,
        method,
        requestBody: body,
        response: response.data,
        status: response.status,
        timestamp: new Date(),
        duration,
        moderationResult: response.moderationResult,
      };

      setTestResults(prev => [result, ...prev]);
      setSnackbar({
        open: true,
        message: `Test completed successfully in ${duration}ms`,
        severity: 'success',
      });
    } catch (error) {
      const duration = Date.now() - startTime;
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      const result: TestResult = {
        id: testId,
        endpoint,
        method,
        requestBody: body,
        response: { error: errorMessage },
        status: 500,
        timestamp: new Date(),
        duration,
      };

      setTestResults(prev => [result, ...prev]);
      setSnackbar({
        open: true,
        message: `Test failed: ${errorMessage}`,
        severity: 'error',
      });
    } finally {
      setCurrentTest(null);
    }
  };

  const simulateAPICall = async (endpoint: string, method: string, body: any) => {
    // Simulate network delay
    await new Promise(resolve => setTimeout(resolve, Math.random() * 1000 + 500));

    // Simulate moderation check
    const content = JSON.stringify(body);
    const isSafe = !content.toLowerCase().includes('spam') && 
                   !content.toLowerCase().includes('hate') &&
                   !content.toLowerCase().includes('buy now') &&
                   !content.toLowerCase().includes('click here');

    if (!isSafe) {
      return {
        status: 403,
        data: {
          error: 'Content violates our community guidelines',
          moderation_result: {
            is_safe: false,
            reason: 'Content contains prohibited terms',
          },
        },
        moderationResult: {
          isSafe: false,
          reason: 'Content contains prohibited terms',
        },
      };
    }

    // Simulate successful response
    return {
      status: 200,
      data: {
        message: 'Content processed successfully',
        id: Math.floor(Math.random() * 10000),
        moderation_result: {
          is_safe: true,
          reason: 'Content passed moderation checks',
        },
      },
      moderationResult: {
        isSafe: true,
        reason: 'Content passed moderation checks',
      },
    };
  };

  const runAutomatedTests = async () => {
    if (!config.realTimeTesting) return;

    setIsRunning(true);
    const tests = [
      { endpoint: '/api/posts', method: 'POST', body: demoContent.safe, type: 'safe_post' },
      { endpoint: '/api/posts', method: 'POST', body: demoContent.unsafe, type: 'unsafe_post' },
      { endpoint: '/api/comments', method: 'POST', body: { postId: 123, content: 'Great article!' }, type: 'safe_comment' },
      { endpoint: '/api/comments', method: 'POST', body: { postId: 123, content: 'I hate this!' }, type: 'unsafe_comment' },
    ];

    for (const test of tests) {
      await handleTest(test.endpoint, test.method, test.body, test.type);
      await new Promise(resolve => setTimeout(resolve, 1000));
    }

    setIsRunning(false);
  };

  const clearResults = () => {
    setTestResults([]);
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setSnackbar({
      open: true,
      message: 'Copied to clipboard!',
      severity: 'success',
    });
  };

  const getStatusIcon = (status: number) => {
    if (status >= 200 && status < 300) return <CheckCircle color="success" />;
    if (status >= 400 && status < 500) return <Block color="error" />;
    return <Warning color="warning" />;
  };

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return 'success';
    if (status >= 400 && status < 500) return 'error';
    return 'warning';
  };

  if (!config.realTimeTesting) {
    return (
      <Alert severity="info">
        <Typography variant="body1">
          Real-time API Testing is disabled. Enable it in the Configuration Panel to test endpoints.
        </Typography>
      </Alert>
    );
  }

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
              🧪 API Testing Console
            </Typography>
            <Box display="flex" gap={2}>
              <Tooltip title={isRunning ? 'Stop Tests' : 'Run Automated Tests'}>
                <IconButton
                  color={isRunning ? 'error' : 'success'}
                  onClick={runAutomatedTests}
                  disabled={isRunning}
                >
                  {isRunning ? <Stop /> : <PlayArrow />}
                </IconButton>
              </Tooltip>
              <Tooltip title="Clear Results">
                <IconButton onClick={clearResults}>
                  <Refresh />
                </IconButton>
              </Tooltip>
            </Box>
          </Box>
          
          <Typography variant="body1" color="text.secondary" paragraph>
            Test your moderation endpoints with real content. See how the system responds to different 
            types of content and monitor moderation decisions in real-time.
          </Typography>

          {isRunning && (
            <Alert severity="info" sx={{ mb: 2 }}>
              <Typography variant="body2">
                🔄 Running automated tests... Please wait
              </Typography>
            </Alert>
          )}
        </Paper>

        <Grid container spacing={3}>
          {/* Post Testing */}
          {config.postTesting && (
            <Grid item xs={12} md={6}>
              <Card elevation={2}>
                <CardHeader
                  title="📝 Post Testing"
                  avatar={<Code />}
                  action={
                    <Button
                      variant="contained"
                      size="small"
                      onClick={() => handleTest('/api/posts', 'POST', postForm, 'post')}
                      disabled={!postForm.title || !postForm.content || currentTest !== null}
                      startIcon={currentTest === 'post' ? <CircularProgress size={16} /> : <Send />}
                    >
                      Test Post
                    </Button>
                  }
                />
                <CardContent>
                  <TextField
                    fullWidth
                    label="Title"
                    value={postForm.title}
                    onChange={(e) => setPostForm(prev => ({ ...prev, title: e.target.value }))}
                    margin="normal"
                  />
                  <TextField
                    fullWidth
                    label="Content"
                    multiline
                    rows={3}
                    value={postForm.content}
                    onChange={(e) => setPostForm(prev => ({ ...prev, content: e.target.value }))}
                    margin="normal"
                  />
                  <TextField
                    fullWidth
                    label="Tags (comma separated)"
                    value={postForm.tags?.join(', ') || ''}
                    onChange={(e) => setPostForm(prev => ({ 
                      ...prev, 
                      tags: e.target.value.split(',').map(tag => tag.trim()).filter(Boolean)
                    }))}
                    margin="normal"
                  />
                  
                  <Box mt={2}>
                    <Typography variant="subtitle2" gutterBottom>
                      Quick Test Examples:
                    </Typography>
                    <Box display="flex" gap={1} flexWrap="wrap">
                      <Chip
                        label="Safe Content"
                        onClick={() => setPostForm(demoContent.safe)}
                        clickable
                        size="small"
                      />
                      <Chip
                        label="Spam Content"
                        onClick={() => setPostForm(demoContent.unsafe)}
                        clickable
                        size="small"
                        color="warning"
                      />
                      <Chip
                        label="Hate Content"
                        onClick={() => setPostForm(demoContent.hate)}
                        clickable
                        size="small"
                        color="error"
                      />
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          )}

          {/* Comment Testing */}
          {config.commentTesting && (
            <Grid item xs={12} md={6}>
              <Card elevation={2}>
                <CardHeader
                  title="💬 Comment Testing"
                  avatar={<Code />}
                  action={
                    <Button
                      variant="contained"
                      size="small"
                      onClick={() => handleTest('/api/comments', 'POST', commentForm, 'comment')}
                      disabled={!commentForm.content || currentTest !== null}
                      startIcon={currentTest === 'comment' ? <CircularProgress size={16} /> : <Send />}
                    >
                      Test Comment
                    </Button>
                  }
                />
                <CardContent>
                  <TextField
                    fullWidth
                    label="Post ID"
                    type="number"
                    value={commentForm.postId}
                    onChange={(e) => setCommentForm(prev => ({ ...prev, postId: parseInt(e.target.value) }))}
                    margin="normal"
                  />
                  <TextField
                    fullWidth
                    label="Comment Content"
                    multiline
                    rows={3}
                    value={commentForm.content}
                    onChange={(e) => setCommentForm(prev => ({ ...prev, content: e.target.value }))}
                    margin="normal"
                  />
                </CardContent>
              </Card>
            </Grid>
          )}

          {/* Profile Testing */}
          {config.profileTesting && (
            <Grid item xs={12} md={6}>
              <Card elevation={2}>
                <CardHeader
                  title="👤 Profile Testing"
                  avatar={<Code />}
                  action={
                    <Button
                      variant="contained"
                      size="small"
                      onClick={() => handleTest('/api/profile', 'PUT', profileForm, 'profile')}
                      disabled={currentTest !== null}
                      startIcon={currentTest === 'profile' ? <CircularProgress size={16} /> : <Send />}
                    >
                      Test Profile
                    </Button>
                  }
                />
                <CardContent>
                  <TextField
                    fullWidth
                    label="Bio"
                    multiline
                    rows={2}
                    value={profileForm.bio}
                    onChange={(e) => setProfileForm(prev => ({ ...prev, bio: e.target.value }))}
                    margin="normal"
                  />
                  <TextField
                    fullWidth
                    label="Website"
                    value={profileForm.website}
                    onChange={(e) => setProfileForm(prev => ({ ...prev, website: e.target.value }))}
                    margin="normal"
                  />
                  <TextField
                    fullWidth
                    label="Location"
                    value={profileForm.location}
                    onChange={(e) => setProfileForm(prev => ({ ...prev, location: e.target.value }))}
                    margin="normal"
                  />
                </CardContent>
              </Card>
            </Grid>
          )}

          {/* Custom Content Testing */}
          {config.customContent && (
            <Grid item xs={12} md={6}>
              <Card elevation={2}>
                <CardHeader
                  title="🔧 Custom Content Testing"
                  avatar={<Code />}
                  action={
                    <Button
                      variant="contained"
                      size="small"
                      onClick={() => handleTest(customForm.endpoint, customForm.method, { content: customForm.content }, 'custom')}
                      disabled={!customForm.content || currentTest !== null}
                      startIcon={currentTest === 'custom' ? <CircularProgress size={16} /> : <Send />}
                    >
                      Test Custom
                    </Button>
                  }
                />
                <CardContent>
                  <TextField
                    fullWidth
                    label="Endpoint"
                    value={customForm.endpoint}
                    onChange={(e) => setCustomForm(prev => ({ ...prev, endpoint: e.target.value }))}
                    margin="normal"
                  />
                  <FormControl fullWidth margin="normal">
                    <InputLabel>Method</InputLabel>
                    <Select
                      value={customForm.method}
                      label="Method"
                      onChange={(e) => setCustomForm(prev => ({ ...prev, method: e.target.value }))}
                    >
                      <MenuItem value="POST">POST</MenuItem>
                      <MenuItem value="PUT">PUT</MenuItem>
                      <MenuItem value="PATCH">PATCH</MenuItem>
                    </Select>
                  </FormControl>
                  <TextField
                    fullWidth
                    label="Content"
                    multiline
                    rows={3}
                    value={customForm.content}
                    onChange={(e) => setCustomForm(prev => ({ ...prev, content: e.target.value }))}
                    margin="normal"
                  />
                </CardContent>
              </Card>
            </Grid>
          )}
        </Grid>

        {/* Test Results */}
        <Card elevation={2} sx={{ mt: 3 }}>
          <CardHeader
            title="📊 Test Results"
            avatar={<History />}
            action={
              <Box display="flex" gap={1}>
                <Typography variant="body2" color="textSecondary">
                  {testResults.length} tests
                </Typography>
                <Button size="small" onClick={clearResults}>
                  Clear All
                </Button>
              </Box>
            }
          />
          <CardContent>
            {testResults.length === 0 ? (
              <Typography variant="body2" color="textSecondary" align="center" sx={{ py: 4 }}>
                No tests run yet. Start testing endpoints to see results here.
              </Typography>
            ) : (
              <List>
                {testResults.map((result, index) => (
                  <React.Fragment key={result.id}>
                    <ListItem>
                      <ListItemIcon>
                        {getStatusIcon(result.status)}
                      </ListItemIcon>
                      <ListItemText
                        primary={
                          <Box display="flex" alignItems="center" gap={1} flexWrap="wrap">
                            <Typography variant="body1" fontWeight="bold">
                              {result.method} {result.endpoint}
                            </Typography>
                            <Chip
                              label={`${result.status}`}
                              color={getStatusColor(result.status) as any}
                              size="small"
                            />
                            <Chip
                              label={`${result.duration}ms`}
                              variant="outlined"
                              size="small"
                            />
                            {result.moderationResult && (
                              <Chip
                                label={result.moderationResult.isSafe ? 'Safe' : 'Blocked'}
                                color={result.moderationResult.isSafe ? 'success' : 'error'}
                                size="small"
                              />
                            )}
                          </Box>
                        }
                        secondary={
                          <Box>
                            <Typography variant="body2" color="textSecondary">
                              {result.timestamp.toLocaleString()}
                            </Typography>
                            <Box mt={1}>
                              <Typography variant="body2" fontWeight="bold">
                                Request:
                              </Typography>
                              <Box
                                component="pre"
                                sx={{
                                  backgroundColor: 'grey.100',
                                  p: 1,
                                  borderRadius: 1,
                                  fontSize: '0.75rem',
                                  overflow: 'auto',
                                  maxWidth: '100%',
                                }}
                              >
                                {JSON.stringify(result.requestBody, null, 2)}
                              </Box>
                            </Box>
                            <Box mt={1}>
                              <Typography variant="body2" fontWeight="bold">
                                Response:
                              </Typography>
                              <Box
                                component="pre"
                                sx={{
                                  backgroundColor: 'grey.100',
                                  p: 1,
                                  borderRadius: 1,
                                  fontSize: '0.75rem',
                                  overflow: 'auto',
                                  maxWidth: '100%',
                                }}
                              >
                                {JSON.stringify(result.response, null, 2)}
                              </Box>
                            </Box>
                            {result.moderationResult?.reason && (
                              <Box mt={1}>
                                <Typography variant="body2" fontWeight="bold">
                                  Moderation Reason:
                                </Typography>
                                <Typography variant="body2" color="textSecondary">
                                  {result.moderationResult.reason}
                                </Typography>
                              </Box>
                            )}
                          </Box>
                        }
                      />
                      <Tooltip title="Copy to clipboard">
                        <IconButton
                          size="small"
                          onClick={() => copyToClipboard(JSON.stringify(result, null, 2))}
                        >
                          <ContentCopy />
                        </IconButton>
                      </Tooltip>
                    </ListItem>
                    {index < testResults.length - 1 && <Divider />}
                  </React.Fragment>
                ))}
              </List>
            )}
          </CardContent>
        </Card>

        <Snackbar
          open={snackbar.open}
          autoHideDuration={6000}
          onClose={() => setSnackbar(prev => ({ ...prev, open: false }))}
        >
          <Alert
            onClose={() => setSnackbar(prev => ({ ...prev, open: false }))}
            severity={snackbar.severity}
            sx={{ width: '100%' }}
          >
            {snackbar.message}
          </Alert>
        </Snackbar>
      </Box>
    </motion.div>
  );
};

export default APITesting;
