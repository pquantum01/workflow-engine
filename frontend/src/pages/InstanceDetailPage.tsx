import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import IconButton from '@mui/material/IconButton';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Popover from '@mui/material/Popover';
import Select from '@mui/material/Select';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import type {
  Process, ProcessInstance, Workflow, WorkflowInstance,
  Task, TaskInstance, StateTransition,
} from '../types';
import { getProcess } from '../api/processes';
import { getProcessInstance, cancelProcessInstance } from '../api/processInstances';
import { listWorkflowInstances } from '../api/workflowInstances';
import { triggerTransition } from '../api/workflowInstances';
import { listTaskInstances, updateTaskStatus, assignTask } from '../api/taskInstances';
import { listStateTransitions } from '../api/stateTransitions';
import { getWorkflow } from '../api/workflows';
import { listTasks } from '../api/tasks';

const instanceStatusColor: Record<string, 'info' | 'success' | 'warning' | 'error'> = {
  running: 'info',
  completed: 'success',
  cancelled: 'warning',
  failed: 'error',
};

const workflowStatusColor: Record<string, 'info' | 'success' | 'warning'> = {
  active: 'info',
  completed: 'success',
  cancelled: 'warning',
};

const taskStatusColor: Record<string, 'default' | 'info' | 'success' | 'error' | 'warning'> = {
  pending: 'default',
  in_progress: 'info',
  completed: 'success',
  failed: 'error',
  skipped: 'warning',
};

interface WorkflowSection {
  workflowInstance: WorkflowInstance;
  workflow: Workflow;
  taskInstances: TaskInstance[];
  taskDefs: Task[];
  transitions: StateTransition[];
}

export default function InstanceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [process, setProcess] = useState<Process | null>(null);
  const [instance, setInstance] = useState<ProcessInstance | null>(null);
  const [sections, setSections] = useState<WorkflowSection[]>([]);
  const [error, setError] = useState('');

  // Assign dialog state
  const [assignAnchor, setAssignAnchor] = useState<HTMLElement | null>(null);
  const [assignTaskId, setAssignTaskId] = useState('');
  const [assignValue, setAssignValue] = useState('');

  // Transition dialog state
  const [transDialogOpen, setTransDialogOpen] = useState(false);
  const [transWiId, setTransWiId] = useState('');
  const [transName, setTransName] = useState('');
  const [transTriggeredBy, setTransTriggeredBy] = useState('');
  const [transTriggerType, setTransTriggerType] = useState('human');

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const inst = await getProcessInstance(id);
      setInstance(inst);
      const proc = await getProcess(inst.process_id);
      setProcess(proc);

      const wis = await listWorkflowInstances(id) ?? [];
      const sects: WorkflowSection[] = [];
      for (const wi of wis) {
        const [wf, tis, trans] = await Promise.all([
          getWorkflow(wi.workflow_id),
          listTaskInstances(wi.id),
          listStateTransitions(wi.id),
        ]);
        const taskDefs = await listTasks(wf.id) ?? [];
        sects.push({
          workflowInstance: wi,
          workflow: wf,
          taskInstances: tis ?? [],
          taskDefs,
          transitions: trans ?? [],
        });
      }
      setSections(sects);
    } catch (err: any) {
      setError(err.message);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  const handleCancel = async () => {
    if (!id) return;
    try {
      await cancelProcessInstance(id);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleTaskAction = async (taskInstanceId: string, status: string) => {
    try {
      await updateTaskStatus(taskInstanceId, { status });
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleAssign = async () => {
    if (!assignTaskId || !assignValue.trim()) return;
    try {
      await assignTask(assignTaskId, { assigned_to: assignValue.trim() });
      setAssignAnchor(null);
      setAssignTaskId('');
      setAssignValue('');
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const openTransitionDialog = (wiId: string, transitionName: string) => {
    setTransWiId(wiId);
    setTransName(transitionName);
    setTransTriggeredBy('');
    setTransTriggerType('human');
    setTransDialogOpen(true);
  };

  const handleTriggerTransition = async () => {
    if (!transWiId || !transName || !transTriggeredBy.trim()) return;
    try {
      await triggerTransition(transWiId, {
        transition_name: transName,
        triggered_by: transTriggeredBy.trim(),
        trigger_type: transTriggerType,
      });
      setTransDialogOpen(false);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  if (!instance || !process) {
    return <Typography>Loading...</Typography>;
  }

  return (
    <Box>
      {/* Header */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
        <IconButton onClick={() => navigate('/instances')}><ArrowBackIcon /></IconButton>
        <Typography variant="h5">{process.name}</Typography>
        <Chip label={instance.status} size="small" color={instanceStatusColor[instance.status] || 'default'} />
        <Typography variant="body2" color="text.secondary">
          Started {new Date(instance.started_at).toLocaleString()}
        </Typography>
        <Box sx={{ flexGrow: 1 }} />
        {instance.status === 'running' && (
          <Button variant="outlined" color="warning" onClick={handleCancel}>Cancel</Button>
        )}
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</Alert>}

      {sections.length === 0 && (
        <Typography color="text.secondary">No workflow instances.</Typography>
      )}

      {sections.map((sect) => {
        const { workflowInstance: wi, workflow: wf, taskInstances: tis, taskDefs, transitions } = sect;
        const taskNameMap: Record<string, string> = {};
        for (const td of taskDefs) {
          taskNameMap[td.id] = td.name;
        }

        // Available transitions from current state
        const availableTransitions = wf.state_machine.transitions.filter(
          (t) => t.from_state === wi.current_state
        );

        return (
          <Accordion key={wi.id} defaultExpanded>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography sx={{ fontWeight: 500 }}>{wf.name}</Typography>
                <Chip label={wi.current_state} size="small" variant="outlined" />
                <Chip label={wi.status} size="small" color={workflowStatusColor[wi.status] || 'default'} />
              </Box>
            </AccordionSummary>
            <AccordionDetails>
              {/* Task Instances */}
              <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>Tasks</Typography>
              {tis.length > 0 ? (
                <Table size="small" sx={{ mb: 2 }}>
                  <TableHead>
                    <TableRow>
                      <TableCell>Task</TableCell>
                      <TableCell>Status</TableCell>
                      <TableCell>Assigned To</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {tis.map((ti) => (
                      <TableRow key={ti.id}>
                        <TableCell>{taskNameMap[ti.task_id] || ti.task_id}</TableCell>
                        <TableCell>
                          <Chip label={ti.status} size="small" color={taskStatusColor[ti.status] || 'default'} />
                        </TableCell>
                        <TableCell>{ti.assigned_to || '—'}</TableCell>
                        <TableCell align="right">
                          <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'flex-end' }}>
                            {ti.status === 'pending' && (
                              <Button size="small" variant="outlined" onClick={() => handleTaskAction(ti.id, 'in_progress')}>
                                Start
                              </Button>
                            )}
                            {ti.status === 'in_progress' && (
                              <>
                                <Button size="small" variant="outlined" color="success" onClick={() => handleTaskAction(ti.id, 'completed')}>
                                  Complete
                                </Button>
                                <Button size="small" variant="outlined" color="error" onClick={() => handleTaskAction(ti.id, 'failed')}>
                                  Fail
                                </Button>
                              </>
                            )}
                            {!['completed', 'failed', 'skipped'].includes(ti.status) && (
                              <Button
                                size="small"
                                variant="text"
                                onClick={(e) => {
                                  setAssignTaskId(ti.id);
                                  setAssignValue(ti.assigned_to || '');
                                  setAssignAnchor(e.currentTarget);
                                }}
                              >
                                Assign
                              </Button>
                            )}
                          </Box>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>No task instances.</Typography>
              )}

              {/* Manual Transitions */}
              {availableTransitions.length > 0 && wi.status === 'active' && (
                <>
                  <Divider sx={{ my: 2 }} />
                  <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>Transitions</Typography>
                  <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', mb: 2 }}>
                    {availableTransitions.map((t) => (
                      <Button
                        key={t.name}
                        variant="outlined"
                        size="small"
                        onClick={() => openTransitionDialog(wi.id, t.name)}
                      >
                        {t.name} &rarr; {t.to_state}
                      </Button>
                    ))}
                  </Box>
                </>
              )}

              {/* Audit Log */}
              <Divider sx={{ my: 2 }} />
              <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>Audit Log</Typography>
              {transitions.length > 0 ? (
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>From</TableCell>
                      <TableCell>To</TableCell>
                      <TableCell>Transition</TableCell>
                      <TableCell>Triggered By</TableCell>
                      <TableCell>Type</TableCell>
                      <TableCell>Time</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {[...transitions].reverse().map((st) => (
                      <TableRow key={st.id}>
                        <TableCell>{st.from_state}</TableCell>
                        <TableCell>{st.to_state}</TableCell>
                        <TableCell>{st.transition_name}</TableCell>
                        <TableCell>{st.triggered_by || '—'}</TableCell>
                        <TableCell>{st.trigger_type}</TableCell>
                        <TableCell>{new Date(st.created_at).toLocaleString()}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Typography variant="body2" color="text.secondary">No transitions recorded yet.</Typography>
              )}
            </AccordionDetails>
          </Accordion>
        );
      })}

      {/* Assign Popover */}
      <Popover
        open={Boolean(assignAnchor)}
        anchorEl={assignAnchor}
        onClose={() => setAssignAnchor(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
      >
        <Box sx={{ p: 2, display: 'flex', gap: 1, alignItems: 'center' }}>
          <TextField
            size="small"
            label="Assign to"
            value={assignValue}
            onChange={(e) => setAssignValue(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter') handleAssign(); }}
          />
          <Button size="small" variant="contained" onClick={handleAssign}>Save</Button>
        </Box>
      </Popover>

      {/* Transition Dialog */}
      <Dialog open={transDialogOpen} onClose={() => setTransDialogOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Trigger: {transName}</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="Triggered by"
            value={transTriggeredBy}
            onChange={(e) => setTransTriggeredBy(e.target.value)}
            sx={{ mt: 1, mb: 2 }}
          />
          <FormControl fullWidth>
            <InputLabel>Trigger type</InputLabel>
            <Select
              value={transTriggerType}
              label="Trigger type"
              onChange={(e) => setTransTriggerType(e.target.value)}
            >
              <MenuItem value="human">Human</MenuItem>
              <MenuItem value="ai_agent">AI Agent</MenuItem>
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setTransDialogOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleTriggerTransition} disabled={!transTriggeredBy.trim()}>
            Trigger
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
