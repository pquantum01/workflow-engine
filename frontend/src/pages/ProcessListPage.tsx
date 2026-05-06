import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import type { Process } from '../types';
import { listProcesses, createProcess, updateProcess, deleteProcess } from '../api/processes';
import ProcessForm from '../components/ProcessForm';

const STATUS_COLORS: Record<string, 'primary' | 'success' | 'default'> = {
  draft: 'primary',
  active: 'success',
  archived: 'default',
};

export default function ProcessListPage() {
  const navigate = useNavigate();
  const [processes, setProcesses] = useState<Process[]>([]);
  const [error, setError] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Process | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Process | null>(null);

  const load = () => {
    listProcesses()
      .then((data) => setProcesses(data ?? []))
      .catch((err) => setError(err.message));
  };

  useEffect(() => { load(); }, []);

  const handleCreate = async (name: string, description: string) => {
    try {
      await createProcess({ name, description });
      setCreateOpen(false);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleEdit = async (name: string, description: string) => {
    if (!editTarget) return;
    try {
      await updateProcess(editTarget.id, { name, description });
      setEditTarget(null);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteProcess(deleteTarget.id);
      setDeleteTarget(null);
      load();
    } catch (err: any) {
      setError(err.message);
      setDeleteTarget(null);
    }
  };

  const handleStatusChange = async (p: Process, status: string) => {
    try {
      await updateProcess(p.id, { status });
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">Processes</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setCreateOpen(true)}>
          New Process
        </Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</Alert>}

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Version</TableCell>
              <TableCell>Created</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {processes.map((p) => (
              <TableRow
                key={p.id}
                hover
                sx={{ cursor: 'pointer' }}
                onClick={() => navigate(`/processes/${p.id}`)}
              >
                <TableCell>{p.name}</TableCell>
                <TableCell>
                  <Chip label={p.status} size="small" color={STATUS_COLORS[p.status] || 'default'} />
                </TableCell>
                <TableCell>{p.version}</TableCell>
                <TableCell>{new Date(p.created_at).toLocaleDateString()}</TableCell>
                <TableCell align="right" onClick={(e) => e.stopPropagation()}>
                  <IconButton size="small" onClick={() => setEditTarget(p)}><EditIcon fontSize="small" /></IconButton>
                  <IconButton size="small" onClick={() => setDeleteTarget(p)}><DeleteIcon fontSize="small" /></IconButton>
                  {p.status === 'draft' && (
                    <Button size="small" onClick={() => handleStatusChange(p, 'active')}>Activate</Button>
                  )}
                  {p.status === 'active' && (
                    <Button size="small" onClick={() => handleStatusChange(p, 'archived')}>Archive</Button>
                  )}
                </TableCell>
              </TableRow>
            ))}
            {processes.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} align="center">No processes yet. Create one to get started.</TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Create dialog */}
      <ProcessForm
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSubmit={handleCreate}
        title="New Process"
      />

      {/* Edit dialog */}
      {editTarget && (
        <ProcessForm
          open={true}
          onClose={() => setEditTarget(null)}
          onSubmit={handleEdit}
          initialName={editTarget.name}
          initialDescription={editTarget.description}
          title="Edit Process"
        />
      )}

      {/* Delete confirm */}
      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogTitle>Delete Process</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Delete "{deleteTarget?.name}"? This cannot be undone. Only draft processes can be deleted.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)}>Cancel</Button>
          <Button onClick={handleDelete} color="error" variant="contained">Delete</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
