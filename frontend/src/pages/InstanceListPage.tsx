import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import type { Process, ProcessInstance } from '../types';
import { listProcesses } from '../api/processes';
import { listProcessInstances, startProcessInstance } from '../api/processInstances';

const statusColor: Record<string, 'info' | 'success' | 'warning' | 'error'> = {
  running: 'info',
  completed: 'success',
  cancelled: 'warning',
  failed: 'error',
};

export default function InstanceListPage() {
  const navigate = useNavigate();
  const [instances, setInstances] = useState<ProcessInstance[]>([]);
  const [processes, setProcesses] = useState<Process[]>([]);
  const [processMap, setProcessMap] = useState<Record<string, string>>({});
  const [error, setError] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedProcessId, setSelectedProcessId] = useState('');

  const load = useCallback(async () => {
    try {
      const [insts, procs] = await Promise.all([
        listProcessInstances(),
        listProcesses(0, 100),
      ]);
      setInstances(insts ?? []);
      setProcesses((procs ?? []).filter((p) => p.status === 'active'));
      const map: Record<string, string> = {};
      for (const p of procs ?? []) {
        map[p.id] = p.name;
      }
      setProcessMap(map);
    } catch (err: any) {
      setError(err.message);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleStart = async () => {
    if (!selectedProcessId) return;
    try {
      await startProcessInstance({ process_id: selectedProcessId });
      setDialogOpen(false);
      setSelectedProcessId('');
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">Process Instances</Typography>
        <Button variant="contained" startIcon={<PlayArrowIcon />} onClick={() => setDialogOpen(true)}>
          Start Process
        </Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</Alert>}

      {instances.length === 0 ? (
        <Typography color="text.secondary">No instances yet. Start a process to create one.</Typography>
      ) : (
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Process</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Started</TableCell>
              <TableCell>Completed</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {instances.map((inst) => (
              <TableRow
                key={inst.id}
                hover
                sx={{ cursor: 'pointer' }}
                onClick={() => navigate(`/instances/${inst.id}`)}
              >
                <TableCell>{processMap[inst.process_id] || inst.process_id}</TableCell>
                <TableCell>
                  <Chip label={inst.status} size="small" color={statusColor[inst.status] || 'default'} />
                </TableCell>
                <TableCell>{new Date(inst.started_at).toLocaleString()}</TableCell>
                <TableCell>{inst.completed_at ? new Date(inst.completed_at).toLocaleString() : '—'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Start Process</DialogTitle>
        <DialogContent>
          <FormControl fullWidth sx={{ mt: 1 }}>
            <InputLabel>Process</InputLabel>
            <Select
              value={selectedProcessId}
              label="Process"
              onChange={(e) => setSelectedProcessId(e.target.value)}
            >
              {processes.map((p) => (
                <MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>
              ))}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleStart} disabled={!selectedProcessId}>Start</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
