import { Routes, Route, Navigate } from "react-router-dom";
import Login from "./pages/Login";
import Layout from "./components/Layout";
import Dashboard from "./pages/Dashboard";
import Chats from "./pages/Chats";
import Calendar from "./pages/Calendar";
import Patients from "./pages/Patients";
import { getUser } from "./api/client";
import Doctors from "./pages/Doctors.jsx";
import RecordsTable from "./pages/RecordsTable";
import WhatsAppWeb from "./pages/WhatsAppWeb";

function Protected({ children }) {
  const user = getUser();
  if (!user) return <Navigate to="/login" replace />;
  return children;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <Protected>
            <Layout />
          </Protected>
        }
      >
          <Route index element={<Dashboard />} />
          <Route path="chats" element={<Chats />} />
          <Route path="chats/:id" element={<Chats />} />
          <Route path="calendar" element={<Calendar />} />
          <Route path="patients" element={<Patients />} />
          <Route path="whatsapp-web" element={<WhatsAppWeb />} />
          <Route path="doctors" element={<Doctors />} />
          <Route path="records" element={<RecordsTable />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
