import { useState } from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import IconButton from '@mui/material/IconButton';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import type { StateMachine, Task } from '../types';

interface StateMachineEditorProps {
  value: StateMachine;
  onChange: (sm: StateMachine) => void;
  tasks: Task[];
}

const RULE_TYPES = ['single', 'all_of', 'any_of', 'n_of_m'];

export default function StateMachineEditor({ value, onChange, tasks }: StateMachineEditorProps) {
  const [newStateName, setNewStateName] = useState('');
  const [newTransName, setNewTransName] = useState('');
  const [newTransFrom, setNewTransFrom] = useState('');
  const [newTransTo, setNewTransTo] = useState('');

  const stateNames = value.states.map((s) => s.name);
  const taskNames = tasks.map((t) => t.name);

  const addState = () => {
    if (!newStateName.trim() || stateNames.includes(newStateName.trim())) return;
    onChange({
      ...value,
      states: [...value.states, { name: newStateName.trim(), description: '' }],
    });
    setNewStateName('');
  };

  const removeState = (name: string) => {
    onChange({
      ...value,
      states: value.states.filter((s) => s.name !== name),
      transitions: value.transitions.filter((t) => t.from_state !== name && t.to_state !== name),
      initial_state: value.initial_state === name ? '' : value.initial_state,
      terminal_states: value.terminal_states.filter((ts) => ts !== name),
    });
  };

  const addTransition = () => {
    if (!newTransName.trim() || !newTransFrom || !newTransTo) return;
    onChange({
      ...value,
      transitions: [...value.transitions, { name: newTransName.trim(), from_state: newTransFrom, to_state: newTransTo }],
    });
    setNewTransName('');
    setNewTransFrom('');
    setNewTransTo('');
  };

  const removeTransition = (idx: number) => {
    onChange({ ...value, transitions: value.transitions.filter((_, i) => i !== idx) });
  };

  const setCompletionRule = (idx: number, type: string, taskRefs: string[], n?: number) => {
    const transitions = [...value.transitions];
    if (!type) {
      transitions[idx] = { ...transitions[idx], completion_rule: undefined };
    } else {
      transitions[idx] = { ...transitions[idx], completion_rule: { type, task_refs: taskRefs, ...(n ? { n } : {}) } };
    }
    onChange({ ...value, transitions });
  };

  return (
    <Box>
      {/* States */}
      <Typography variant="subtitle1" sx={{ mt: 2, mb: 1, fontWeight: 600 }}>States</Typography>
      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 1 }}>
        {value.states.map((s) => (
          <Chip
            key={s.name}
            label={s.name}
            color={s.name === value.initial_state ? 'primary' : value.terminal_states.includes(s.name) ? 'success' : 'default'}
            onDelete={() => removeState(s.name)}
          />
        ))}
      </Box>
      <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
        <TextField size="small" label="State name" value={newStateName} onChange={(e) => setNewStateName(e.target.value)} />
        <Button variant="outlined" startIcon={<AddIcon />} onClick={addState} disabled={!newStateName.trim()}>Add</Button>
      </Box>

      {/* Initial / Terminal */}
      <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
        <FormControl size="small" sx={{ minWidth: 150 }}>
          <InputLabel>Initial State</InputLabel>
          <Select
            value={value.initial_state}
            label="Initial State"
            onChange={(e) => onChange({ ...value, initial_state: e.target.value })}
          >
            {stateNames.map((n) => <MenuItem key={n} value={n}>{n}</MenuItem>)}
          </Select>
        </FormControl>
        <FormControl size="small" sx={{ minWidth: 200 }}>
          <InputLabel>Terminal States</InputLabel>
          <Select
            multiple
            value={value.terminal_states}
            label="Terminal States"
            onChange={(e) => onChange({ ...value, terminal_states: e.target.value as string[] })}
            renderValue={(selected) => selected.join(', ')}
          >
            {stateNames.map((n) => <MenuItem key={n} value={n}>{n}</MenuItem>)}
          </Select>
        </FormControl>
      </Box>

      <Divider sx={{ my: 2 }} />

      {/* Transitions */}
      <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>Transitions</Typography>
      {value.transitions.map((t, idx) => (
        <Box key={idx} sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1, flexWrap: 'wrap' }}>
          <Chip label={`${t.name}: ${t.from_state} → ${t.to_state}`} variant="outlined" />
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>Rule</InputLabel>
            <Select
              value={t.completion_rule?.type || ''}
              label="Rule"
              onChange={(e) => setCompletionRule(idx, e.target.value, t.completion_rule?.task_refs || [], t.completion_rule?.n)}
            >
              <MenuItem value="">None</MenuItem>
              {RULE_TYPES.map((rt) => <MenuItem key={rt} value={rt}>{rt}</MenuItem>)}
            </Select>
          </FormControl>
          {t.completion_rule && (
            <FormControl size="small" sx={{ minWidth: 180 }}>
              <InputLabel>Task Refs</InputLabel>
              <Select
                multiple
                value={t.completion_rule.task_refs}
                label="Task Refs"
                onChange={(e) => setCompletionRule(idx, t.completion_rule!.type, e.target.value as string[], t.completion_rule!.n)}
                renderValue={(selected) => selected.join(', ')}
              >
                {taskNames.map((n) => <MenuItem key={n} value={n}>{n}</MenuItem>)}
              </Select>
            </FormControl>
          )}
          {t.completion_rule?.type === 'n_of_m' && (
            <TextField
              size="small"
              type="number"
              label="N"
              sx={{ width: 70 }}
              value={t.completion_rule.n || ''}
              onChange={(e) => setCompletionRule(idx, t.completion_rule!.type, t.completion_rule!.task_refs, parseInt(e.target.value) || 0)}
            />
          )}
          <IconButton size="small" onClick={() => removeTransition(idx)}><DeleteIcon fontSize="small" /></IconButton>
        </Box>
      ))}
      <Box sx={{ display: 'flex', gap: 1, mt: 1 }}>
        <TextField size="small" label="Transition name" value={newTransName} onChange={(e) => setNewTransName(e.target.value)} />
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>From</InputLabel>
          <Select value={newTransFrom} label="From" onChange={(e) => setNewTransFrom(e.target.value)}>
            {stateNames.map((n) => <MenuItem key={n} value={n}>{n}</MenuItem>)}
          </Select>
        </FormControl>
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>To</InputLabel>
          <Select value={newTransTo} label="To" onChange={(e) => setNewTransTo(e.target.value)}>
            {stateNames.map((n) => <MenuItem key={n} value={n}>{n}</MenuItem>)}
          </Select>
        </FormControl>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={addTransition} disabled={!newTransName.trim() || !newTransFrom || !newTransTo}>Add</Button>
      </Box>
    </Box>
  );
}
