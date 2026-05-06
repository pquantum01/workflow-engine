import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import SaveIcon from '@mui/icons-material/Save';
import type { Process, Workflow, Task, StateMachine } from '../types';
import { getProcess } from '../api/processes';
import { listWorkflows, createWorkflow, updateWorkflow, deleteWorkflow } from '../api/workflows';
import { listTasks, createTask, deleteTask } from '../api/tasks';
import WorkflowForm from '../components/WorkflowForm';
import TaskForm from '../components/TaskForm';
import StateMachineEditor from '../components/StateMachineEditor';

export default function ProcessDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [process, setProcess] = useState<Process | null>(null);
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [tasksByWorkflow, setTasksByWorkflow] = useState<Record<string, Task[]>>({});
  const [error, setError] = useState('');
  const [workflowFormOpen, setWorkflowFormOpen] = useState(false);
  const [taskFormTarget, setTaskFormTarget] = useState<string | null>(null); // workflow ID
  const [editingSM, setEditingSM] = useState<Record<string, StateMachine>>({}); // workflow ID -> edited SM

  const load = useCallback(async () => {
    if (!id) return;
    try {
      const p = await getProcess(id);
      setProcess(p);
      const wfs = await listWorkflows(id) ?? [];
      setWorkflows(wfs);
      const tasksMap: Record<string, Task[]> = {};
      for (const wf of wfs) {
        tasksMap[wf.id] = await listTasks(wf.id) ?? [];
      }
      setTasksByWorkflow(tasksMap);
    } catch (err: any) {
      setError(err.message);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  const handleCreateWorkflow = async (name: string, description: string, sm: StateMachine) => {
    if (!id) return;
    try {
      await createWorkflow(id, { name, description, state_machine: sm });
      setWorkflowFormOpen(false);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleDeleteWorkflow = async (wfId: string) => {
    try {
      await deleteWorkflow(wfId);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleCreateTask = async (name: string, description: string, taskType: string) => {
    if (!taskFormTarget) return;
    try {
      await createTask(taskFormTarget, { name, description, task_type: taskType });
      setTaskFormTarget(null);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleDeleteTask = async (taskId: string) => {
    try {
      await deleteTask(taskId);
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleSaveStateMachine = async (wfId: string) => {
    const sm = editingSM[wfId];
    if (!sm) return;
    try {
      await updateWorkflow(wfId, { state_machine: sm });
      setEditingSM((prev) => { const next = { ...prev }; delete next[wfId]; return next; });
      load();
    } catch (err: any) {
      setError(err.message);
    }
  };

  if (!process) {
    return <Typography>Loading...</Typography>;
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
        <IconButton onClick={() => navigate('/')}><ArrowBackIcon /></IconButton>
        <Typography variant="h5">{process.name}</Typography>
        <Chip label={process.status} size="small" color={process.status === 'active' ? 'success' : process.status === 'draft' ? 'primary' : 'default'} />
      </Box>
      {process.description && <Typography color="text.secondary" sx={{ mb: 2 }}>{process.description}</Typography>}

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</Alert>}

      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
        <Typography variant="h6">Workflows</Typography>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => setWorkflowFormOpen(true)} disabled={process.status === 'archived'}>
          Add Workflow
        </Button>
      </Box>

      {workflows.length === 0 && <Typography color="text.secondary" sx={{ mb: 2 }}>No workflows yet.</Typography>}

      {workflows.map((wf) => {
        const tasks = tasksByWorkflow[wf.id] || [];
        const currentSM = editingSM[wf.id] || wf.state_machine;
        const smDirty = !!editingSM[wf.id];

        return (
          <Accordion key={wf.id} defaultExpanded>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, width: '100%' }}>
                <Typography sx={{ fontWeight: 500 }}>{wf.name}</Typography>
                <Typography variant="body2" color="text.secondary">
                  ({wf.state_machine.states.length} states, {wf.state_machine.transitions.length} transitions)
                </Typography>
                <Box sx={{ flexGrow: 1 }} />
                <IconButton size="small" onClick={(e) => { e.stopPropagation(); handleDeleteWorkflow(wf.id); }}>
                  <DeleteIcon fontSize="small" />
                </IconButton>
              </Box>
            </AccordionSummary>
            <AccordionDetails>
              {/* Tasks */}
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>Tasks</Typography>
                <Button size="small" startIcon={<AddIcon />} onClick={() => setTaskFormTarget(wf.id)} disabled={process.status === 'archived'}>
                  Add Task
                </Button>
              </Box>
              {tasks.length > 0 ? (
                <Table size="small" sx={{ mb: 2 }}>
                  <TableHead>
                    <TableRow>
                      <TableCell>Name</TableCell>
                      <TableCell>Type</TableCell>
                      <TableCell align="right">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {tasks.map((t) => (
                      <TableRow key={t.id}>
                        <TableCell>{t.name}</TableCell>
                        <TableCell><Chip label={t.task_type} size="small" variant="outlined" /></TableCell>
                        <TableCell align="right">
                          <IconButton size="small" onClick={() => handleDeleteTask(t.id)}><DeleteIcon fontSize="small" /></IconButton>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>No tasks yet.</Typography>
              )}

              <Divider sx={{ my: 2 }} />

              {/* State Machine */}
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>State Machine</Typography>
                {smDirty && (
                  <Button size="small" variant="contained" startIcon={<SaveIcon />} onClick={() => handleSaveStateMachine(wf.id)}>
                    Save Changes
                  </Button>
                )}
              </Box>
              <StateMachineEditor
                value={currentSM}
                onChange={(sm) => setEditingSM((prev) => ({ ...prev, [wf.id]: sm }))}
                tasks={tasks}
              />
            </AccordionDetails>
          </Accordion>
        );
      })}

      {/* Workflow form dialog */}
      <WorkflowForm
        open={workflowFormOpen}
        onClose={() => setWorkflowFormOpen(false)}
        onSubmit={handleCreateWorkflow}
        title="New Workflow"
      />

      {/* Task form dialog */}
      {taskFormTarget && (
        <TaskForm
          open={true}
          onClose={() => setTaskFormTarget(null)}
          onSubmit={handleCreateTask}
          title="New Task"
        />
      )}
    </Box>
  );
}
