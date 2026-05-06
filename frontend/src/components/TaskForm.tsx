import { useState } from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';

interface TaskFormProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (name: string, description: string, taskType: string) => void;
  initialName?: string;
  initialDescription?: string;
  initialTaskType?: string;
  title: string;
}

const TASK_TYPES = ['manual', 'automated', 'approval'];

export default function TaskForm({ open, onClose, onSubmit, initialName = '', initialDescription = '', initialTaskType = 'manual', title }: TaskFormProps) {
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);
  const [taskType, setTaskType] = useState(initialTaskType);

  const handleSubmit = () => {
    if (!name.trim()) return;
    onSubmit(name.trim(), description.trim(), taskType);
    setName('');
    setDescription('');
    setTaskType('manual');
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <TextField
          autoFocus
          required
          label="Name"
          fullWidth
          margin="normal"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <TextField
          label="Description"
          fullWidth
          margin="normal"
          multiline
          rows={3}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <TextField
          select
          label="Task Type"
          fullWidth
          margin="normal"
          value={taskType}
          onChange={(e) => setTaskType(e.target.value)}
        >
          {TASK_TYPES.map((t) => (
            <MenuItem key={t} value={t}>{t}</MenuItem>
          ))}
        </TextField>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSubmit} variant="contained" disabled={!name.trim()}>
          Save
        </Button>
      </DialogActions>
    </Dialog>
  );
}
