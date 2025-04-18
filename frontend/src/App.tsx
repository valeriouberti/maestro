import { BrowserRouter as Router, Route, Routes, Navigate } from 'react-router-dom';
import ClusterList from './components/ClusterList';
import TopicList from './components/TopicList';
import TopicDetails from './components/TopicDetails';
import ConsumerGroupList from './components/ConsumerGroupList';
import ConsumerGroupDetails from './components/ConsumerGroupDetails';
import TopicForm from './components/TopicForm';
import Sidebar from './components/Sidebar';
import './index.css';

function App() {
  return (
    <Router>
      <div className="flex h-screen bg-white text-gray-800">
        <Sidebar />
        <div className="flex-1 p-6 ml-64 overflow-y-auto">
          <div className="max-w-5xl mx-auto">
            <Routes>
              <Route path="/" element={<Navigate to="/clusters" replace />} />
              <Route path="/clusters" element={<ClusterList />} />
              <Route path="/topics" element={<TopicList />} />
              <Route path="/topics/:topicName" element={<TopicDetails />} />
              <Route path="/consumer-groups" element={<ConsumerGroupList />} />
              <Route path="/consumer-groups/:groupId" element={<ConsumerGroupDetails />} />
              <Route path="/topics/create" element={<TopicForm />} />
            </Routes>
          </div>
        </div>
      </div>
    </Router>
  );
}

export default App;