import { useState } from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import type { StateMachine } from '../types';

interface WorkflowFormProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (name: string, description: string, stateMachine: StateMachine) => void;
  initialName?: string;
  initialDescription?: string;
  title: string;
}

const defaultStateMachine: StateMachine = {
  states: [{ name: 'open', description: '' }, { name: 'closed', description: '' }],
  transitions: [{ name: 'close', from_state: 'open', to_state: 'closed' }],
  initial_state: 'open',
  terminal_states: ['closed'],
};

export default function WorkflowForm({ open, onClose, onSubmit, initialName = '', initialDescription = '', title }: WorkflowFormProps) {
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);

  const handleSubmit = () => {
    if (!name.trim()) return;
    onSubmit(name.trim(), description.trim(), defaultStateMachine);
    setName('');
    setDescription('');
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
