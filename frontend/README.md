# 🎛️ API Gateway Moderation System - React Frontend

A beautiful, interactive React frontend for the API Gateway Moderation System with a **configuration-driven approach** that gives you full control over your experience.

## ✨ Features

### 🎨 **Interactive Configuration Panel**
- **UI Style & Theme**: Choose from Modern, Colorful, Dark, or Light themes
- **Color Schemes**: Blue, Green, Purple, or Custom colors
- **Layout Options**: Compact, Spacious, Sidebar, or Top navigation
- **Real-time Preview**: See changes instantly as you configure

### 🚀 **Modular Features**
- **Real-time Testing**: Live API endpoint testing with instant results
- **Moderation Dashboard**: Visual statistics and performance metrics
- **Content Forms**: Interactive forms for testing different content types
- **Statistics & Analytics**: Detailed metrics and performance data
- **User Management**: Advanced authentication and user management

### 🔐 **Authentication Options**
- **Demo Mode**: Use pre-generated tokens for testing
- **Login Form**: Interactive login for real authentication
- **User Registration**: Create new user accounts

### 📝 **Content Testing**
- **Post Testing**: Test blog post content moderation
- **Comment Testing**: Test comment content moderation
- **Profile Testing**: Test user profile content moderation
- **Custom Content**: Test arbitrary content with custom validation

### 🌐 **Deployment & API Settings**
- **Port Configuration**: Choose your preferred frontend port
- **API Endpoint**: Configure backend URL
- **Proxy Settings**: Handle CORS and proxy configuration

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ 
- npm or yarn
- Go backend running (see main README)

### Installation

1. **Navigate to the frontend directory:**
   ```bash
   cd API_Gateway/frontend
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Start the development server:**
   ```bash
   npm start
   ```

4. **Open your browser:**
   Navigate to `http://localhost:3000`

## 🎯 How to Use

### 1. **Configuration Panel** (`/`)
- Start here to customize your experience
- Choose your preferred UI style and features
- Enable/disable functionality as needed
- Save your configuration for future use

### 2. **Moderation Dashboard** (`/dashboard`)
- View real-time moderation statistics
- Monitor system performance
- Track blocked content and reasons
- Enable live updates for dynamic data

### 3. **API Testing Console** (`/testing`)
- Test moderation endpoints with real content
- Use pre-built examples or create custom content
- Run automated test suites
- View detailed test results and moderation decisions

## 🎨 Customization Options

### **Quick Theme Presets**
- **Modern Blue**: Clean, professional interface
- **Colorful Purple**: Creative, vibrant design
- **Demo Mode Only**: Simplified testing experience
- **Minimal Features**: Streamlined functionality

### **Advanced Customization**
- Custom color palettes
- Layout adjustments
- Feature toggles
- Performance settings

## 🔧 Configuration Examples

### **For Developers**
```typescript
// Enable all features for development
{
  theme: 'modern',
  colorScheme: 'blue',
  realTimeTesting: true,
  moderationDashboard: true,
  contentForms: true,
  statistics: true,
  userManagement: true
}
```

### **For Testing**
```typescript
// Minimal setup for testing
{
  theme: 'light',
  colorScheme: 'green',
  realTimeTesting: true,
  contentForms: true,
  moderationDashboard: false,
  statistics: false,
  userManagement: false
}
```

### **For Production**
```typescript
// Production-ready configuration
{
  theme: 'modern',
  colorScheme: 'blue',
  realTimeTesting: false,
  moderationDashboard: true,
  contentForms: false,
  statistics: true,
  userManagement: true
}
```

## 🧪 Testing Your Moderation System

### **Quick Test Examples**
1. **Safe Content**: "Go is a wonderful programming language..."
2. **Spam Content**: "Buy now! Limited time offer..."
3. **Hate Content**: "I hate this content so much..."

### **Automated Testing**
- Click the "Run Automated Tests" button
- Watch as the system processes different content types
- See moderation decisions in real-time
- Review detailed test results

## 📱 Responsive Design

- **Desktop**: Full-featured interface with side-by-side panels
- **Tablet**: Optimized layout for medium screens
- **Mobile**: Touch-friendly interface with stacked components

## 🎭 Animations & Interactions

- **Smooth Transitions**: Framer Motion animations
- **Hover Effects**: Interactive card and button animations
- **Loading States**: Visual feedback during operations
- **Real-time Updates**: Live data refresh with smooth transitions

## 🔌 API Integration

The frontend is designed to work with your Go backend:

- **Default Endpoint**: `http://localhost:8080`
- **Moderation Endpoints**: `/api/posts`, `/api/comments`, `/api/profile`
- **Public Endpoints**: `/public/feedback`
- **Demo Endpoint**: `/demo`

## 🛠️ Development

### **Available Scripts**
```bash
npm start          # Start development server
npm run build      # Build for production
npm test           # Run tests
npm run eject      # Eject from Create React App
```

### **Project Structure**
```
src/
├── components/           # React components
│   ├── ConfigurationPanel.tsx    # Main configuration interface
│   ├── ModerationDashboard.tsx   # Statistics and metrics
│   └── APITesting.tsx           # API testing console
├── contexts/            # React contexts
│   └── ConfigContext.tsx        # Configuration state management
├── App.tsx             # Main application component
└── index.css           # Global styles and animations
```

### **Key Technologies**
- **React 19** with TypeScript
- **Material-UI (MUI)** for components
- **Framer Motion** for animations
- **React Router** for navigation
- **Context API** for state management

## 🎨 Theme System

### **Built-in Themes**
- **Modern**: Clean, professional design
- **Colorful**: Vibrant, creative interface
- **Dark**: Dark mode for low-light environments
- **Light**: Bright, accessible design

### **Custom Colors**
- Primary, secondary, and accent color pickers
- Real-time color preview
- Persistent color schemes

## 🚀 Performance Features

- **Lazy Loading**: Components load on demand
- **Optimized Rendering**: Efficient React patterns
- **Smooth Animations**: 60fps animations with Framer Motion
- **Responsive Design**: Optimized for all screen sizes

## 🔒 Security Features

- **JWT Authentication**: Secure token-based auth
- **Content Validation**: Input sanitization and validation
- **CORS Handling**: Proper cross-origin request handling
- **Secure Storage**: Local storage for configuration

## 📊 Monitoring & Analytics

- **Real-time Metrics**: Live performance data
- **Content Statistics**: Moderation success rates
- **System Health**: Performance and availability monitoring
- **User Activity**: Track testing and usage patterns

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the main repository for details.

## 🆘 Support

- **Documentation**: Check the main README.md
- **Issues**: Report bugs via GitHub Issues
- **Discussions**: Join community discussions
- **Examples**: See the examples directory

## 🎉 What's Next?

Your configuration-driven moderation system is ready! 

1. **Start the backend**: `go run examples/moderation_integration.go`
2. **Start the frontend**: `npm start`
3. **Configure your experience**: Use the Configuration Panel
4. **Test your system**: Use the API Testing Console
5. **Monitor performance**: Check the Moderation Dashboard

Enjoy your fully customizable moderation system! 🚀✨
