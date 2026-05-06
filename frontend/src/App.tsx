import { BrowserRouter, Routes, Route } from 'react-router-dom';
import CssBaseline from '@mui/material/CssBaseline';
import Layout from './components/Layout';
import ProcessListPage from './pages/ProcessListPage';
import ProcessDetailPage from './pages/ProcessDetailPage';
import InstanceListPage from './pages/InstanceListPage';
import InstanceDetailPage from './pages/InstanceDetailPage';
import NotFoundPage from './pages/NotFoundPage';

export default function App() {
  return (
    <BrowserRouter>
      <CssBaseline />
      <Layout>
        <Routes>
          <Route path="/" element={<ProcessListPage />} />
          <Route path="/processes/:id" element={<ProcessDetailPage />} />
          <Route path="/instances" element={<InstanceListPage />} />
          <Route path="/instances/:id" element={<InstanceDetailPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}
