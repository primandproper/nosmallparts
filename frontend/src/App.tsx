import { BrowserRouter, Route, Routes } from 'react-router';

import { AuthProvider } from './lib/auth';
import Home from './routes/Home';
import Login from './routes/Login';
import Register from './routes/Register';
import ScriptDetail from './routes/ScriptDetail';
import ScriptEditor from './routes/ScriptEditor';
import ScriptsLibrary from './routes/ScriptsLibrary';
import StageScreen from './routes/StageScreen';

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/scripts" element={<ScriptsLibrary />} />
          <Route path="/scripts/new" element={<ScriptEditor />} />
          <Route path="/scripts/:id" element={<ScriptDetail />} />
          <Route path="/stage/:stageId" element={<StageScreen />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}
