from .api_backend import ApiBackend
from .change_volume_voice_call import ChangeVolumeVoiceCall
from .event_finish import EventFinish
from .get_participants import GetParticipants
from .http_request_manager import HTTPRequestManager
from .join_voice_call import JoinVoiceCall
from .leave_voice_call import LeaveVoiceCall
from .load_chat_call import LoadChatCall
from .prepare_run import PrepareRun
from .start_web_app import StartWebApp
from .update_call_data import UpdateCallData


class WebSocket(
    ApiBackend,
    ChangeVolumeVoiceCall,
    EventFinish,
    GetParticipants,
    HTTPRequestManager,
    JoinVoiceCall,
    LeaveVoiceCall,
    LoadChatCall,
    PrepareRun,
    StartWebApp,
    UpdateCallData,
):
    pass
